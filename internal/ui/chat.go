package ui

import (
	"fmt"
	"slices"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thero-sgit/glyn/internal/ui/library"
	"github.com/thero-sgit/glyn/internal/ui/submodels/components"
)

// -- CHAT CENTRE CONTROL
type chatCentre struct {
	sessionName       string
	isAgentWorking    bool
	isChatClear       bool
	history           []chatItem
	currentUserPrompt userPrompt
	currentAgentRes   agentResponse
	sendingPrompt     bool
	fileSelector      components.FileSelector
	filesForContext   []string
}

func newChatCentre() chatCentre {
	return chatCentre{
		isAgentWorking: false,
		isChatClear:    true,
		fileSelector:   components.NewFileSelector(),
	}
}

func (c *chatCentre) prettyHistory() []string {
	h := []string{}
	for _, item := range c.history {
		h = append(h, item.getPretty())
	}

	return h
}

func (c *chatCentre) updatedWidths(width int) {
	var updatedHistory []chatItem
	for _, v := range c.history {
		updatedHistory = append(updatedHistory, v.updated(width))
	}

	c.history = updatedHistory
}

func (c *chatCentre) addFileForContext(path string) {
	if !slices.Contains(c.filesForContext, path) {
		c.filesForContext = append(c.filesForContext, path)
	}
}

func (c *chatCentre) removeFileForContext(path string) {
	c.filesForContext = slices.DeleteFunc(c.filesForContext, func(s string) bool {
		return s == path
	})
}

// -- CHAT ITEMS CONTROL
type chatItem interface {
	getPretty() string
	updated(width int) chatItem
}

type userPrompt struct {
	index     int
	width     int
	message   string
	waveIndex int
	pretty    string
}

func (up userPrompt) getPretty() string {
	return up.pretty
}

func (up userPrompt) updated(width int) chatItem {
	updated := newUserPrompt(up.message, width)
	updated.index = up.index
	updated.sent()

	return updated
}

func newUserPrompt(prompt string, width int) userPrompt {
	width = width - 6
	pretty := lipgloss.NewStyle().
		MarginBottom(1).
		Width(width).
		PaddingRight(1).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				library.HighlighedStyle.Render(" you $ "),
				lipgloss.NewStyle().PaddingLeft(1).Render(prompt),
			),
		)

	return userPrompt{
		width:     width,
		message:   prompt,
		waveIndex: 0,
		pretty:    pretty,
	}
}

func (up *userPrompt) sent() {
	p := library.HighlighedStyle.Render(" you $ ") + lipgloss.NewStyle().PaddingLeft(1).Render(up.message)

	up.pretty = lipgloss.NewStyle().
		MarginBottom(1).
		Width(up.width).
		PaddingRight(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				p,
				lipgloss.NewStyle().Foreground(lipgloss.Color("#02e696")).PaddingRight(2).Render("\u2713"),
			),
		)
}

func (up *userPrompt) err() {
	p := library.HighlighedStyle.Render(" you $ ") + lipgloss.NewStyle().Strikethrough(true).PaddingLeft(1).Render(up.message)

	ed := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().MarginRight(4).Render(
			library.SubtleStyle.Render("failed to send"),
		),
		library.RetryButton(),
	)

	up.pretty = lipgloss.NewStyle().
		MarginBottom(1).
		Width(up.width).
		PaddingRight(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				p,
				ed,
			),
		)
}

type agentResponse struct {
	index           int
	width           int
	agentBgActivity *agentBackgroundActivityLabel
	reasoningBuffer string
	responseBuffer  string
	pretty          string
	showReasoning   bool
}

func (a agentResponse) updated(width int) chatItem {
	updated := newAgentResponse(width)
	updated.responseBuffer = a.responseBuffer
	updated.agentBgActivity = a.agentBgActivity
	updated.reasoningBuffer = a.reasoningBuffer

	return updated
}

func (a agentResponse) getPretty() string {
	content := a.responseBuffer
	if a.showReasoning {
		reasoningContent := library.SubtleStyle.Width(a.width - 2).Render(a.reasoningBuffer)
		content = lipgloss.JoinVertical(
			lipgloss.Top,
			reasoningContent,
			"\n",
			content,
		)
	}

	return lipgloss.NewStyle().Width(a.width).MarginBottom(1).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			a.agentBgActivity.prettyString,
			content,
		),
	)
}

func newAgentResponse(width int) agentResponse {
	agentBgActivity := newAgentBackgroundActivity("Thinking")

	pretty := lipgloss.NewStyle().Width(width).MarginBottom(1).Render(lipgloss.JoinVertical(
		lipgloss.Top,
		agentBgActivity.prettyString,
		lipgloss.NewStyle().
			Render(""),
	),
	)

	return agentResponse{
		width:           width,
		agentBgActivity: &agentBgActivity,
		pretty:          pretty,
	}
}

type agentBackgroundActivityLabel struct {
	index        int
	prefixLabel  string
	rawString    string
	loaderIndex  int
	loaderFrames []string
	loader       string
	prettyString string
	doneStyle    lipgloss.Style
	intrptStyle  lipgloss.Style
	loaderStyle  lipgloss.Style
}

func newAgentBackgroundActivity(activity string) agentBackgroundActivityLabel {
	prefixLabel := library.BwLabelStyle.Render(" glyn $ ")

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	loader := frames[0]

	doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Italic(true)
	loaderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E47753")).Bold(true)
	intrptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Bold(true)

	prettyString := fmt.Sprintf(
		"%s %s %s...",
		prefixLabel,
		loaderStyle.Render(loader),
		activity,
	)

	return agentBackgroundActivityLabel{
		prefixLabel:  prefixLabel,
		rawString:    activity,
		loaderIndex:  0,
		loaderFrames: frames,
		loader:       loader,
		prettyString: prettyString,
		doneStyle:    doneStyle,
		intrptStyle:  intrptStyle,
		loaderStyle:  loaderStyle,
	}
}

func (abal *agentBackgroundActivityLabel) animate() {
	abal.loaderIndex = (abal.loaderIndex + 1) % len(abal.loaderFrames)

	abal.prettyString = fmt.Sprintf(
		"%s %s %s...",
		abal.prefixLabel,
		abal.loaderStyle.Render(abal.loaderFrames[abal.loaderIndex]),
		abal.rawString,
	)
}

func (abal *agentBackgroundActivityLabel) done(onErr bool) {
	var concl string
	if onErr {
		concl = abal.intrptStyle.Render("*")
	} else {
		concl = abal.doneStyle.Render("Thought process") + " >"
	}

	prettyString := fmt.Sprintf(
		"%s %s",
		abal.prefixLabel,
		concl,
	)

	abal.prettyString = zone.Mark("thought-process-btn", prettyString)
}
