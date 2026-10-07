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
// TODO:
// 1 tab switching
// 2 text input
// 3 posts and repliees
// 4 backend & sqlite DB
// 5 nest app + that thingy ynk what i mena why i am i typing this only ill read this

package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	width  int
	height int
}

// tab switcching logic

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	}
	return m, nil
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

func header(text string, w, h int) string {
	return lipgloss.NewStyle().
		Bold(true).
		Italic(true).
		Background(lipgloss.Color("#1d2021")). // workeds??? prob should do the ciolro beteter but wtv
		Foreground(lipgloss.Color("#a89984")).
		Width(w).
		Height(h).
		AlignHorizontal(lipgloss.Center). // Copy ts over and delete ( commented out ) in pane()? ^^
		AlignVertical(lipgloss.Center).
		Render(text)
}

func (m model) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.NewView("loading...")
	}

	topHeight := 1
	bottomHeight := 4
	midHeight := m.height - topHeight - bottomHeight
	leftWidth := m.width * 3 / 4
	rightWidth := m.width - leftWidth

	top := pane(m.width, topHeight, "#1d2021", header("board.tgyt.dev", m.width, topHeight))
	left := pane(leftWidth, midHeight, "#282828", "Post")
	right := pane(rightWidth, midHeight, "#3c3836", "Replies")
	bottom := pane(m.width, bottomHeight, "#1d2021", "Post Maker")

	middle := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	content := lipgloss.JoinVertical(lipgloss.Left, top, middle, bottom)

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
