package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	gossh "golang.org/x/crypto/ssh"
)

func TestHashPubKey(t *testing.T) {
	// Test nil key
	if hash := hashPubKey(nil); hash != "" {
		t.Fatalf("expected empty string for nil public key, got %q", hash)
	}

	// Generate test ed25519 key
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	sshPub, err := gossh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("failed to convert to ssh public key: %v", err)
	}

	expectedHash := sha256.Sum256(sshPub.Marshal())
	expectedHex := hex.EncodeToString(expectedHash[:])

	actualHex := hashPubKey(sshPub.(ssh.PublicKey))
	if actualHex != expectedHex {
		t.Fatalf("expected hash %q, got %q", expectedHex, actualHex)
	}

	if len(actualHex) != 64 {
		t.Fatalf("expected 64 characters for sha256 hex string, got %d", len(actualHex))
	}
}

func TestNewModelWithUserID(t *testing.T) {
	db, err := openDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer db.Close()

	testUserID := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	m := newModel(db, testUserID)

	if m.userID != testUserID {
		t.Fatalf("expected model userID to be %q, got %q", testUserID, m.userID)
	}
}

func TestSetUsername(t *testing.T) {
	db, err := openDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer db.Close()

	const userID = "username-test-user"
	if err := setUsername(db, userID, "first-name"); err != nil {
		t.Fatalf("setUsername failed: %v", err)
	}
	if err := setUsername(db, userID, "updated-name"); err != nil {
		t.Fatalf("updating username failed: %v", err)
	}

	username, err := getUsername(db, userID)
	if err != nil {
		t.Fatalf("getUsername failed: %v", err)
	}
	if username != "updated-name" {
		t.Fatalf("expected updated username, got %q", username)
	}
}

func TestUsernameEntrySavesUsernameInsteadOfPost(t *testing.T) {
	db, err := openDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer db.Close()

	const userID = "new-user"
	m := newModel(db, userID)
	m.postMaker.SetValue("  codex  ")

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 's', Mod: tea.ModCtrl}))
	updatedModel := updated.(model)

	username, err := getUsername(db, userID)
	if err != nil {
		t.Fatalf("getUsername failed: %v", err)
	}
	if username != "codex" {
		t.Fatalf("expected saved username %q, got %q", "codex", username)
	}
	if updatedModel.settingUsername || updatedModel.typing {
		t.Fatal("expected username entry to finish after saving")
	}
	posts, err := getPosts(db, 0, 10)
	if err != nil {
		t.Fatalf("getPosts failed: %v", err)
	}
	if len(posts) != 0 {
		t.Fatalf("expected no post to be created while setting a username, got %d", len(posts))
	}
}

func TestSSHServerPubkeySession(t *testing.T) {
	db, err := openDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	var receivedUserID string
	connected := make(chan struct{})

	l, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer l.Close()

	customHandler := func(sess ssh.Session) {
		receivedUserID = hashPubKey(sess.PublicKey())
		close(connected)
		sess.Exit(0)
	}

	s, err := wish.NewServer(
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		}),
		func(srv *ssh.Server) error {
			srv.Handler = customHandler
			return nil
		},
	)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	go func() {
		_ = s.Serve(l)
	}()
	defer s.Close()

	// Client key
	clientPub, clientPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate client key: %v", err)
	}
	clientSigner, err := gossh.NewSignerFromKey(clientPriv)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	sshClientPub, err := gossh.NewPublicKey(clientPub)
	if err != nil {
		t.Fatalf("failed to get client pubkey: %v", err)
	}
	expectedHash := sha256.Sum256(sshClientPub.Marshal())
	expectedUserID := hex.EncodeToString(expectedHash[:])

	clientConfig := &gossh.ClientConfig{
		User:            "testuser",
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(clientSigner)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}

	client, err := gossh.Dial("tcp", l.Addr().String(), clientConfig)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	defer session.Close()

	_ = session.Shell()

	select {
	case <-connected:
		if receivedUserID != expectedUserID {
			t.Fatalf("expected userID %q, got %q", expectedUserID, receivedUserID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for session")
	}
}

func TestMakeAndGetPosts(t *testing.T) {
	db, err := openDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	userID := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	err = makePost(db, userID, "Hello World Post", 0)
	if err != nil {
		t.Fatalf("makePost failed: %v", err)
	}

	posts, err := getPosts(db, 0, 10)
	if err != nil {
		t.Fatalf("getPosts failed: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if posts[0].content != "Hello World Post" {
		t.Fatalf("unexpected content: %s", posts[0].content)
	}
	if posts[0].userId != userID {
		t.Fatalf("unexpected userId: %s", posts[0].userId)
	}

	// Test reply
	rootID := posts[0].postId
	err = makePost(db, userID, "This is a reply", rootID)
	if err != nil {
		t.Fatalf("makePost reply failed: %v", err)
	}

	replies, err := getPosts(db, rootID, 10)
	if err != nil {
		t.Fatalf("getPosts replies failed: %v", err)
	}
	if len(replies) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(replies))
	}
	if replies[0].content != "This is a reply" {
		t.Fatalf("unexpected reply content: %s", replies[0].content)
	}
}

func TestToggleLike(t *testing.T) {
	db, err := openDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	userID := "1122334455667788990011223344556677889900112233445566778899001122"
	_ = makePost(db, userID, "Post to like", 0)
	posts, _ := getPosts(db, 0, 1)
	postID := posts[0].postId

	// Toggle like ON
	err = toggleLike(db, postID, userID)
	if err != nil {
		t.Fatalf("toggleLike on failed: %v", err)
	}

	posts, _ = getPosts(db, 0, 1)
	if len(posts[0].likes) != 1 {
		t.Fatalf("expected 1 like, got %d", len(posts[0].likes))
	}

	// Toggle like OFF
	err = toggleLike(db, postID, userID)
	if err != nil {
		t.Fatalf("toggleLike off failed: %v", err)
	}
	posts, _ = getPosts(db, 0, 1)
	if len(posts[0].likes) != 0 {
		t.Fatalf("expected 0 likes, got %d", len(posts[0].likes))
	}
}
