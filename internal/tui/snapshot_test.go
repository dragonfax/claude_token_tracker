package tui

import (
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestMain(m *testing.M) {
	os.Setenv("TZ", "UTC")
	lipgloss.SetColorProfile(termenv.ANSI256)
	lipgloss.SetHasDarkBackground(true)
	os.Exit(m.Run())
}

// Watch view tests

func TestWatchSnapshot_Empty(t *testing.T) {
	model := watchModel{
		width:  testWidth,
		height: testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "watch_empty", got)
}

func TestWatchSnapshot_WithEntries(t *testing.T) {
	model := watchModel{
		entries: standardEntries(),
		width:   testWidth,
		height:  testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "watch_with_entries", got)
}

func TestWatchSnapshot_ManyEntries(t *testing.T) {
	model := watchModel{
		entries: makeNEntries(50),
		width:   testWidth,
		height:  testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "watch_many_entries", got)
}

func TestWatchSnapshot_ShowSubOff(t *testing.T) {
	model := watchModel{
		entries: standardEntries(),
		showSub: false,
		width:   testWidth,
		height:  testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "watch_show_sub_off", got)
}

func TestWatchSnapshot_ShowSubOn(t *testing.T) {
	model := watchModel{
		entries: standardEntries(),
		showSub: true,
		width:   testWidth,
		height:  testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "watch_show_sub_on", got)
}

func TestWatchSnapshot_WithSummaries(t *testing.T) {
	model := watchModel{
		entries: entriesWithSummaries(),
		showSub: true,
		width:   testWidth,
		height:  testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "watch_with_summaries", got)
}

// Log view tests

func TestLogSnapshot_Empty(t *testing.T) {
	model := logModel{
		loaded: true,
		width:  testWidth,
		height: testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "log_empty", got)
}

func TestLogSnapshot_WithEntries(t *testing.T) {
	model := logModel{
		entries: standardEntries(),
		showSub: true,
		loaded:  true,
		offset:  0,
		width:   testWidth,
		height:  testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "log_with_entries", got)
}

func TestLogSnapshot_ScrollTop(t *testing.T) {
	model := logModel{
		entries: makeNEntries(50),
		showSub: true,
		loaded:  true,
		offset:  0,
		width:   testWidth,
		height:  testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "log_scroll_top", got)
}

func TestLogSnapshot_ScrollMid(t *testing.T) {
	model := logModel{
		entries: makeNEntries(50),
		showSub: true,
		loaded:  true,
		offset:  20,
		width:   testWidth,
		height:  testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "log_scroll_mid", got)
}

func TestLogSnapshot_ErrorsOnly(t *testing.T) {
	model := logModel{
		entries:    standardEntries(),
		errorsOnly: true,
		loaded:     true,
		width:      testWidth,
		height:     testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "log_errors_only", got)
}

func TestLogSnapshot_Window24h(t *testing.T) {
	model := logModel{
		entries:   standardEntries(),
		showSub:   true,
		loaded:    true,
		windowIdx: 0,
		width:     testWidth,
		height:    testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "log_window_24h", got)
}

func TestLogSnapshot_Window7d(t *testing.T) {
	model := logModel{
		entries:   standardEntries(),
		showSub:   true,
		loaded:    true,
		windowIdx: 2,
		width:     testWidth,
		height:    testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "log_window_7d", got)
}

func TestLogSnapshot_WithSummaries(t *testing.T) {
	model := logModel{
		entries: entriesWithSummaries(),
		showSub: true,
		loaded:  true,
		offset:  0,
		width:   testWidth,
		height:  testHeight,
	}
	got := model.View()
	assertFitsInTerminal(t, got, testHeight)
	assertGolden(t, "log_with_summaries", got)
}
