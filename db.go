package main

import (
	"database/sql"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

// i wrote the final stuff on my phoen in obsidina on the bus so i copied it here

type user struct {
	userID   string // yoinked from ssh pub key
	username string // shii lowk gotta make a keybind for that or maybe like user idk
}

type post struct {
	postId    int64
	userId    string
	content   string
	rootPost  int64 // 0 if not reply
	timestamp int64
	likes     []string // NOTE: ts is a nucha user stringts UNIUQE or whover you spell that bs
	username  string
}

// i ahve some expierence iwth sqlite from a discord bot i did a tad back
// im lowk tryna copy the logic from ^^^ + some examples off the internet

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	_, _ = db.Exec("PRAGMA journal_mode=WAL;")
	_, _ = db.Exec("PRAGMA busy_timeout=5000;")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			postId INTEGER PRIMARY KEY AUTOINCREMENT,
			userId TEXT NOT NULL, 
			content TEXT NOT NULL,
			rootPost INTEGER NOT NULL DEFAULT 0,
			timestamp INTEGER NOT NULL,
			username TEXT
		);
	`)
	if err != nil {
		return nil, err
	}

	// Ensure username column exists if messages table pre-dated it
	_, _ = db.Exec("ALTER TABLE messages ADD COLUMN username TEXT;")

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS likes (
			postId INTEGER NOT NULL,
			userId TEXT NOT NULL, 
			PRIMARY KEY (postId, userId)
		);
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			userId TEXT PRIMARY KEY, 
			username TEXT NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func makePost(db *sql.DB, userID, content string, rootPost int64) error {
	username, _ := getUsername(db, userID)
	_, err := db.Exec(
		"INSERT INTO messages (userId, content, rootPost, timestamp, username) VALUES (?, ?, ?, ?, ?)",
		userID, content, rootPost, time.Now().Unix(), username,
	)
	return err
}

func getPosts(db *sql.DB, rootPost int64, limit int) ([]post, error) {
	rows, err := db.Query(
		"SELECT m.postId, m.userId, m.content, m.rootPost, m.timestamp, COALESCE(NULLIF(m.username, ''), u.username, '') FROM messages m LEFT JOIN users u ON m.userId = u.userId WHERE m.rootPost = ? ORDER BY m.postId DESC LIMIT ?",
		rootPost, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []post
	for rows.Next() {
		var p post
		if err := rows.Scan(&p.postId, &p.userId, &p.content, &p.rootPost, &p.timestamp, &p.username); err != nil {
			continue
		}
		posts = append(posts, p)
	}
	_ = rows.Close()

	for i := range posts {
		likeRows, err := db.Query("SELECT userId FROM likes WHERE postId = ?", posts[i].postId)
		if err != nil {
			continue
		}
		for likeRows.Next() {
			var likeUser string
			if err := likeRows.Scan(&likeUser); err == nil {
				posts[i].likes = append(posts[i].likes, likeUser)
			}
		}
		_ = likeRows.Close()
	}

	return posts, nil
}

func toggleLike(db *sql.DB, postId int64, userId string) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE postId = ? AND userId = ?", postId, userId).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		_, err = db.Exec("DELETE FROM likes WHERE postId = ? AND userId = ?", postId, userId)
		return err
	}
	_, err = db.Exec("INSERT INTO likes (postId, userId) VALUES (?, ?)", postId, userId)
	return err
}

// scrapped notifs

func getNotifs(db *sql.DB) int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func getUsername(db *sql.DB, userId string) (string, error) {
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE userId = ?", userId).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

func setUsername(db *sql.DB, userId, username string) error {
	_, err := db.Exec(
		"INSERT INTO users (userId, username) VALUES (?, ?) ON CONFLICT(userId) DO UPDATE SET username = ?",
		userId, username, username,
	)
	return err
}
