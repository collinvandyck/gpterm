package ui

import tea "github.com/charmbracelet/bubbletea"

type commands []tea.Cmd

func (c *commands) Add(cmd tea.Cmd) {
	*c = append(*c, cmd)
}

func (c commands) SequenceWith(cmds ...tea.Cmd) tea.Cmd {
	cmds = append(cmds, c...)
	if len(cmds) == 0 {
		return nil
	}
	return tea.Sequence(cmds...)
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
