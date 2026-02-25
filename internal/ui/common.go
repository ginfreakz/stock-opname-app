package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// focusableTable wraps a Table and intercepts key events so keyboard shortcuts
// work even after clicking inside the table.
type focusableTable struct {
	widget.BaseWidget
	table      *widget.Table
	onTypedKey func(*fyne.KeyEvent)
	focused    bool
}

func newFocusableTable(table *widget.Table, onTypedKey func(*fyne.KeyEvent)) *focusableTable {
	ft := &focusableTable{
		table:      table,
		onTypedKey: onTypedKey,
	}
	ft.ExtendBaseWidget(ft)
	return ft
}

func (ft *focusableTable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ft.table)
}

func (ft *focusableTable) FocusGained()     { ft.focused = true }
func (ft *focusableTable) FocusLost()       { ft.focused = false }
func (ft *focusableTable) TypedRune(r rune) {}

func (ft *focusableTable) TypedKey(ev *fyne.KeyEvent) {
	if ft.onTypedKey != nil {
		ft.onTypedKey(ev)
	}
}

func (ft *focusableTable) KeyDown(ev *fyne.KeyEvent) {}
func (ft *focusableTable) KeyUp(ev *fyne.KeyEvent)   {}

var _ fyne.Focusable = (*focusableTable)(nil)
