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
	"fmt"
	"os"
	"strconv"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// tab switcching logic

const (
	topHeight    = 1
	bottomHeight = 4
)

type focus int

const (
	focusPost focus = iota
	focusReplies
)

// ill jus append to the haeder instead of colors

func (f focus) name() string {
	switch f {
	case focusPost:
		return "Post"
	case focusReplies:
		return "Replies"
	}
	return ""
}

type model struct {
	width     int
	height    int
	focus     focus
	typing    bool // NOTE: use for displaiying the inpuit box instead of having a static one ... cooler ig
	postMaker textarea.Model
}

// post input thingy

func newModel() model {
	postMakerIniter := textarea.New()
	postMakerIniter.ShowLineNumbers = false
	postMakerIniter.Prompt = ""
	postMakerIniter.Placeholder = "Type in your post / reply (Markdown is supported)"
	return model{postMaker: postMakerIniter}
}

// cool REAL header name abstraction-inator 3000

func (m model) realHeaderNameinator3000() string {
	if m.typing {
		if m.focus == focusReplies {
			return "Replying to PLACEHOLDER" // HACK: Hardcoded until sqlite db setup - auto gen title or like make a req
		}
		return "Viewing PLACEHOLDER" // llook above dipshit
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
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+tab":
			if !m.typing {
				m.focus = (m.focus + 1) % 3
			}
		case "ctrl+n":
			if m.typing {
				return m, nil
			}
			m.typing = true
			return m, m.postMaker.Focus()
		case "ctrl+e":
			if m.typing {
				m.typing = false
				m.postMaker.Blur()
				m.postMaker.Reset()
			}
			return m, nil
		case "ctrl+s": // TODO: SQLite stuff
			if m.typing {
				postContent := m.postMaker.Value()
				if m.focus == focusReplies {
					_ = postContent
				} else {
					_ = postContent
				}
				m.typing = false
				m.postMaker.Blur()
				m.postMaker.Reset()
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

func getNotifs() int {
	return 1 // HACK: Hardcoded - design like stuff in db - fetch all posts with their used id and getch like lenghts and replies
}

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

	midHeight := m.height - topHeight - bottomHeight
	leftWidth := m.width * 3 / 4
	rightWidth := m.width - leftWidth

	top := pane(m.width, topHeight, "#1d2021", header(getNotifsText(getNotifs()), "board.tgyt.dev", m.realHeaderNameinator3000(), m.width, topHeight))
	left := pane(leftWidth, midHeight, "#282828", "Post")
	right := pane(rightWidth, midHeight, "#3c3836", "Replies")
	bottomText := " ctrl+n: new post"
	if m.focus == focusReplies {
		bottomText = " ctrl+n: reply to this post"
	}
	if m.typing {
		bottomText = m.postMaker.View()
	}
	bottom := pane(m.width, bottomHeight, "#1d2021", bottomText)

	middle := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	content := lipgloss.JoinVertical(lipgloss.Left, top, middle, bottom)

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
