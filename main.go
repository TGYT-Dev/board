// DESIGN OUTLINE
// account - ssh pub key (sha256 hash)
// user from computer naem? or manually enter (prob latter)
// ---
// post struct - sqlite probs
// POSTID: sequential
// USERID: sha256 hash - OP
// CONTENT: Md?? maybe custom idk basic italics underline bold and 3 headers - lip gloss or wtv has like cool uhm
// ISREPLY: bool - not nested *yet*
// └> PARENTID - OP'S POST ID
// TIMESTAMP
// ---
// user struct
// USERID: sha256 of pub key
// USER: str - choosen
//
// 4 panes = lipgloss
// upher is titile
// content | replies
// message box for new posts
//
// one pane?
// ctrl+n = new message
// ctrl+r = reply (only active when viewing a original or no cuz uhm itll be split betewwn the message and replies ( 3/4 and 1/4) )
// down (scroll message unless on replies then scroll replies) and up ynk
//
// FINAL DATABASE SCHEME
// ---
// USER
// 	USERID = sha256 pub
//	USERNAME = chosen
// ---
// MESSAGE
// 	POSTID ( Sequential )
//  USERID ( From OP )
//  CONTENT ( MD )
//  ROOTPOST ( 0 if original post )
// 	LIKES [ USERIDs ]
// ---
// When loading replies fetch all items with ROOTPOST == POSTID
// those will be loaded into the side panel ( right )
// ---
// Dev Path??
// 1 tab switching - done
// 2 text input - through into nothing rn ^^^
// 3 posts and repliees
// 4 backend & sqlite DB
// 5 nest app + that thingy ynk what i mena why i am i typing this only ill read this
// set up grabing pub key

package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/bubbletea"
)

// siziing

const (
	topHeight      = 1
	bottomHeight   = 4
	initMsgsLoaded = 50
)

// tab stufff

type focus int

const (
	focusPost focus = iota
	focusReplies
)

// ill jus append to the haeder instead of colors - tf did i mean here?

func (f focus) name() string {
	switch f {
	case focusPost:
		return "Post"
	case focusReplies:
		return "Replies"
	}
	return ""
}

// user stuff - vscode is really cool for formatting, my lovely nvim would never ]:<

type model struct {
	width           int
	height          int
	focus           focus
	typing          bool // NOTE: use for displaiying the inpuit box instead of having a static one ... cooler ig
	postMaker       textarea.Model
	db              *sql.DB
	userID          string
	posts           []post
	replies         []post
	cursor          int
	offset          int
	settingUsername bool
	username        string
}

// post input thingy

func newModel(db *sql.DB, userID string) model {
	postMakerIniter := textarea.New()
	postMakerIniter.ShowLineNumbers = false
	postMakerIniter.Prompt = ""
	existingUsername, err := getUsername(db, userID)

	settingUsername := false

	if err != nil || existingUsername == "" {
		settingUsername = true
		postMakerIniter.Placeholder = "Enter your chosen username and press ctrl+enter"
	} else {
		postMakerIniter.Placeholder = "Type in your post / reply (Markdown is supported)"
	}

	bg := lipgloss.Color("#1d2021")
	fg := lipgloss.Color("#ebdbb2")
	dim := lipgloss.Color("#928374")
	accent := lipgloss.Color("#fabd2f")

	s := postMakerIniter.Styles()
	s.Focused.Base = lipgloss.NewStyle().Background(bg)
	s.Focused.Text = lipgloss.NewStyle().Background(bg).Foreground(fg)
	s.Focused.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("#282828")).Foreground(fg)
	s.Focused.Placeholder = lipgloss.NewStyle().Background(bg).Foreground(dim).Italic(true)
	s.Focused.EndOfBuffer = lipgloss.NewStyle().Background(bg)
	s.Cursor.Color = accent
	postMakerIniter.SetStyles(s)

	initPosts, _ := getPosts(db, 0, initMsgsLoaded)

	return model{
		postMaker:       postMakerIniter,
		db:              db,
		userID:          userID,
		username:        existingUsername,
		posts:           initPosts,
		settingUsername: settingUsername,
		typing:          settingUsername,
	}
}

// hashPubKey returns the stable user ID associated with an SSH public key.
// Marshal is the canonical SSH wire representation, so the result does not
// depend on how the key was formatted in an authorized_keys file.
func hashPubKey(key ssh.PublicKey) string {
	if key == nil {
		return ""
	}

	hash := sha256.Sum256(key.Marshal())
	return hex.EncodeToString(hash[:])
}

// cool REAL header name abstraction-inator 3000

func (m model) realHeaderNameinator3000() string {
	if m.typing {
		if m.focus == focusReplies && len(m.posts) > 0 {
			return fmt.Sprintf("Replying to Post #[%d]", m.posts[m.cursor].postId)
		}
		return "Writing New Post"
	}
	return m.focus.name()
}

// other stuff

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.postMaker.SetWidth(m.width)
		m.postMaker.SetHeight(bottomHeight)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j":
			if !m.typing && m.cursor < len(m.posts)-1 {
				m.cursor++
			}
		case "k":
			if !m.typing && m.cursor > 0 {
				m.cursor--
			}
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "shift+tab":
			if !m.typing {
				m.focus = (m.focus + 1) % 2
			}
			return m, nil
		case "ctrl+n":
			if m.typing {
				return m, nil
			}
			m.typing = true
			return m, m.postMaker.Focus()
		case "ctrl+e":
			if m.typing && !m.settingUsername {
				m.typing = false
				m.postMaker.Blur()
				m.postMaker.Reset()
			}
			return m, nil
		case "ctrl+enter", "ctrl+s":
			if m.typing {
				// Ctrl+S is commonly consumed by SSH terminal flow control, so
				// Ctrl+Enter submits a username without turning Enter into a
				// submit key for the multiline textarea.
				if msg.String() == "ctrl+enter" && !m.settingUsername {
					break
				}
				postContent := strings.TrimSpace(m.postMaker.Value())
				if m.settingUsername {
					if postContent == "" || setUsername(m.db, m.userID, postContent) != nil {
						return m, nil
					}

					m.username = postContent
					m.settingUsername = false
					m.postMaker.Placeholder = "Type in your post / reply (Markdown is supported)"
					m.typing = false
					m.postMaker.Blur()
					m.postMaker.Reset()
					return m, nil
				}

				if postContent != "" {
					var rootPost int64 = 0
					if m.focus == focusReplies && len(m.posts) > 0 {
						rootPost = m.posts[m.cursor].postId
					}

					// Save post using db.go helper
					_ = makePost(m.db, m.userID, postContent, rootPost)

					// Refresh the post list immediately
					if rootPost == 0 {
						m.posts, _ = getPosts(m.db, 0, 50)
					} else {
						m.replies, _ = getPosts(m.db, rootPost, 50)
					}
				}
				m.typing = false
				m.postMaker.Blur()
				m.postMaker.Reset()
			}
			return m, nil
		case "ctrl+l": // Like the currently selected post
			if !m.typing && len(m.posts) > 0 {
				selectedPost := m.posts[m.cursor]
				_ = toggleLike(m.db, selectedPost.postId, m.userID)

				// Refresh posts to update the like counts
				m.posts, _ = getPosts(m.db, 0, 50)
			}
			return m, nil

		} // NOTE: Didnt need the reply... streamline posting with ctrl n for all depends on tab selected now
	}
	var cmd tea.Cmd
	if m.typing {
		m.postMaker, cmd = m.postMaker.Update(msg)
	}
	return m, cmd
}

func pane(w, h int, bg string, text string) string {
	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color("#ffffff")).
		// AlignHorizontal(lipgloss.Center). // Revert to basic tut answer
		// AlignVertical(lipgloss.Center).   // this centers everthhing else tho
		Render(text)
}

func header(left, title, right string, w, h int) string {
	base := lipgloss.NewStyle().
		Bold(true).
		Italic(true).
		Background(lipgloss.Color("#1d2021")).
		Foreground(lipgloss.Color("#a89984")).
		Height(h)

	sideW := w / 4
	centerW := w - 2*sideW

	leftPart := base.Width(sideW).AlignHorizontal(lipgloss.Left).Render(" " + left)
	center := base.Width(centerW).AlignHorizontal(lipgloss.Center).Render(title)
	rightPart := base.Width(sideW).AlignHorizontal(lipgloss.Right).Render(right + " ")

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPart, center, rightPart)
}

// notifs bs scrapped - too complex and takes too long - shipping soon

func getNotifsText(notifCount int) string {
	// do user get notifs stuff
	var notifStr string
	if notifCount < 10 && notifCount != 1 {
		notifStr = strconv.Itoa(notifCount) + " Notifications"
	} else if notifCount == 1 {
		notifStr = "1 Notification"
	} else if notifCount > 9 {
		notifStr = "9+ Notifications" // truncate bullshit blah blah blah
	} else {
		notifStr = "No Notifications"
	}
	// maybe swap to switch - im too lazy
	return lipgloss.NewStyle().
		Bold(true).
		Italic(true).
		Background(lipgloss.Color("#1d2021")).
		Foreground(lipgloss.Color("#fb4934")).
		Render(notifStr)

}

func (m model) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.NewView("loading...")
	}

	bh := 0
	if m.typing {
		bh = bottomHeight
	}
	midHeight := m.height - topHeight - bh
	leftWidth := m.width * 3 / 4
	rightWidth := m.width - leftWidth

	// notifCount := getNotifs(m.db)

	// Format posts for the left pane
	leftContent := "No posts yet (Something could be wrong?). Press ctrl+n to write one!"
	if len(m.posts) > 0 {
		leftContent = ""
		for i, p := range m.posts {
			cursor := "  "
			if i == m.cursor && m.focus == focusPost {
				cursor = "> " // Show a cursor marker on the active post
			}
			author := p.username
			if author == "" {
				author = p.userId[:8]
			}
			leftContent += fmt.Sprintf("%s[%d] %s (♥ %d)\n%s\n\n", cursor, p.postId, author, len(p.likes), p.content)
		}
	}

	rightContent := "Select a post to view replies"
	if len(m.posts) > 0 {
		selectedPostID := m.posts[m.cursor].postId
		replies, _ := getPosts(m.db, selectedPostID, 20)
		if len(replies) > 0 {
			rightContent = ""
			for _, r := range replies {
				rightContent += fmt.Sprintf("[%d] %s:\n%s\n\n", r.postId, r.userId[:8], r.content)
			}
		} else {
			rightContent = "No replies yet."
		}
	}

	top := pane(m.width, topHeight, "#1d2021", header("", "board.tgyt.dev", m.realHeaderNameinator3000(), m.width, topHeight))
	left := pane(leftWidth, midHeight, "#282828", leftContent)
	right := pane(rightWidth, midHeight, "#3c3836", rightContent)

	middle := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	parts := []string{top, middle}
	if m.typing {
		parts = append(parts, pane(m.width, bottomHeight, "#1d2021", m.postMaker.View()))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "board.db"
	}
	db, err := openDB(dbPath)
	if err != nil {
		fmt.Println("database error:", err)
		os.Exit(1)
	}
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "2222"
	}
	address := net.JoinHostPort(os.Getenv("HOST"), port)

	hostKeyPath := os.Getenv("KEY_PATH")
	if hostKeyPath == "" {
		hostKeyPath = ".ssh/id_ed25519"
	}

	server, err := wish.NewServer(
		wish.WithAddress(address),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithPublicKeyAuth(func(_ ssh.Context, _ ssh.PublicKey) bool { return true }),
		wish.WithMiddleware(bubbletea.Middleware(func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
			return newModel(db, hashPubKey(sess.PublicKey())), nil
		})),
	)
	if err != nil {
		fmt.Println("SSH server error:", err)
		os.Exit(1)
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Println("SSH server error:", err)
		os.Exit(1)
	}
}
