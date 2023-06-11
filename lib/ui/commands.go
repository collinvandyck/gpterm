package ui

import tea "github.com/charmbracelet/bubbletea"

type commands []tea.Cmd

func (c *commands) Add(cmds ...tea.Cmd) {
	cmds = c.removeNils(cmds)
	*c = append(*c, cmds...)
}

func (c *commands) Insert(cmds ...tea.Cmd) {
	cmds = c.removeNils(cmds)
	*c = append(cmds, *c...)
}

func (c *commands) removeNils(cmds []tea.Cmd) (res []tea.Cmd) {
	res = make([]tea.Cmd, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd == nil {
			continue
		}
		res = append(res, cmd)
	}
	return
}

func (c commands) Sequence() tea.Cmd {
	if len(c) == 0 {
		return nil
	}
	return tea.Sequence(c...)
}

func (c commands) BatchWith(cmds ...tea.Cmd) tea.Cmd {
	cmds = append(cmds, c...)
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (c *commands) Update(model tea.Model, msg tea.Msg) tea.Model {
	model, cmd := model.Update(msg)
	c.Add(cmd)
	return model
}
