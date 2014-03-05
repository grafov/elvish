package edit

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// Line editor builtins.
// These are not exposed to the user in anyway yet. Ideally, they should
// reside in a dedicated namespace and callable by users, e.g.
// le:kill-line-f.

type editorAction int

const (
	noAction editorAction = iota
	changeMode
	changeModeAndReprocess
	exitReadLine
)

type leReturn struct {
	action         editorAction
	newMode        bufferMode
	readLineReturn LineRead
}

type leBuiltin func(ed *Editor, k Key) *leReturn

var leBuiltins = map[string]leBuiltin{
	// Command and insert mode
	"start-insert":    startInsert,
	"start-command":   startCommand,
	"kill-line-b":     killLineB,
	"kill-line-f":     killLineF,
	"kill-rune-b":     killRuneB,
	"move-dot-b":      moveDotB,
	"move-dot-f":      moveDotF,
	"return-line":     returnLine,
	"return-eof":      returnEOF,
	"default-command": defaultCommand,
	"default-insert":  defaultInsert,

	// Completion mode
	"start-completion":   startCompletion,
	"cancel-completion":  cancelCompletion,
	"select-cand-b":      selectCandB,
	"select-cand-f":      selectCandF,
	"select-cand-col-b":  selectCandColB,
	"select-cand-col-f":  selectCandColF,
	"cycle-cand-f":       cycleCandF,
	"default-completion": defaultCompletion,

	// Navigation mode
	"start-navigation":   startNavigation,
	"default-navigation": defaultNavigation,

	// History mode
	"start-history":    startHistory,
	"cancel-history":   cancelHistory,
	"select-history-b": selectHistoryB,
	"select-history-f": selectHistoryF,
	"default-history":  defaultHistory,
}

func startInsert(ed *Editor, k Key) *leReturn {
	return &leReturn{action: changeMode, newMode: modeInsert}
}

func defaultCommand(ed *Editor, k Key) *leReturn {
	ed.pushTip(fmt.Sprintf("Unbound: %s", k))
	return nil
}

func startCommand(ed *Editor, k Key) *leReturn {
	return &leReturn{action: changeMode, newMode: modeCommand}
}

func killLineB(ed *Editor, k Key) *leReturn {
	ed.line = ed.line[ed.dot:]
	ed.dot = 0
	return nil
}

func killLineF(ed *Editor, k Key) *leReturn {
	ed.line = ed.line[:ed.dot]
	return nil
}

func killRuneB(ed *Editor, k Key) *leReturn {
	if ed.dot > 0 {
		_, w := utf8.DecodeLastRuneInString(ed.line[:ed.dot])
		ed.line = ed.line[:ed.dot-w] + ed.line[ed.dot:]
		ed.dot -= w
	} else {
		ed.beep()
	}
	return nil
}

func moveDotB(ed *Editor, k Key) *leReturn {
	_, w := utf8.DecodeLastRuneInString(ed.line[:ed.dot])
	ed.dot -= w
	return nil
}

func moveDotF(ed *Editor, k Key) *leReturn {
	_, w := utf8.DecodeRuneInString(ed.line[ed.dot:])
	ed.dot += w
	return nil
}

func returnLine(ed *Editor, k Key) *leReturn {
	return &leReturn{action: exitReadLine, readLineReturn: LineRead{Line: ed.line}}
}

func returnEOF(ed *Editor, k Key) *leReturn {
	if len(ed.line) == 0 {
		return &leReturn{action: exitReadLine, readLineReturn: LineRead{EOF: true}}
	}
	return nil
}

func selectCandB(ed *Editor, k Key) *leReturn {
	ed.completion.prev(false)
	return nil
}

func selectCandF(ed *Editor, k Key) *leReturn {
	ed.completion.next(false)
	return nil
}

func selectCandColB(ed *Editor, k Key) *leReturn {
	if c := ed.completion.current - ed.completionLines; c >= 0 {
		ed.completion.current = c
	}
	return nil
}

func selectCandColF(ed *Editor, k Key) *leReturn {
	if c := ed.completion.current + ed.completionLines; c < len(ed.completion.candidates) {
		ed.completion.current = c
	}
	return nil
}

func cycleCandF(ed *Editor, k Key) *leReturn {
	ed.completion.next(true)
	return nil
}

func cancelCompletion(ed *Editor, k Key) *leReturn {
	ed.completion = nil
	ed.mode = modeInsert
	return nil
}

func defaultInsert(ed *Editor, k Key) *leReturn {
	if k.Mod == 0 && k.rune > 0 && unicode.IsGraphic(k.rune) {
		ed.line = ed.line[:ed.dot] + string(k.rune) + ed.line[ed.dot:]
		ed.dot += utf8.RuneLen(k.rune)
	} else {
		ed.pushTip(fmt.Sprintf("Unbound: %s", k))
	}
	return nil
}

func defaultCompletion(ed *Editor, k Key) *leReturn {
	ed.acceptCompletion()
	return &leReturn{action: changeModeAndReprocess, newMode: modeInsert}
}

func startNavigation(ed *Editor, k Key) *leReturn {
	ed.mode = modeNavigation
	ed.navigation = newNavigation()
	return &leReturn{}
}

func defaultNavigation(ed *Editor, k Key) *leReturn {
	ed.mode = modeInsert
	ed.navigation = nil
	return &leReturn{}
}

func startHistory(ed *Editor, k Key) *leReturn {
	ed.history.saved = ed.line
	ed.history.prefix = ed.line[:ed.dot]
	ed.history.current = len(ed.history.items)
	if ed.history.prev() {
		ed.mode = modeHistory
	} else {
		ed.pushTip("no matching history item")
	}
	return nil
}

func cancelHistory(ed *Editor, k Key) *leReturn {
	ed.mode = modeInsert
	return nil
}

func selectHistoryB(ed *Editor, k Key) *leReturn {
	ed.history.prev()
	return nil
}

func selectHistoryF(ed *Editor, k Key) *leReturn {
	ed.history.next()
	return nil
}

func defaultHistory(ed *Editor, k Key) *leReturn {
	ed.acceptHistory()
	return &leReturn{action: changeModeAndReprocess, newMode: modeInsert}
}
