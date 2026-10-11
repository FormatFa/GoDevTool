package main

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"indigo6a.online/gokit/internal/portscan"
)

func runes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

var sampleItems = []portscan.Item{
	{LocalHost: "0.0.0.0", Port: "135", PID: "1044", Process: "System"},
	{LocalHost: "127.0.0.1", Port: "8080", PID: "12345", Process: "node.exe"},
}

func loadedModel(t *testing.T) model {
	t.Helper()
	m := initialModel()
	m2, _ := m.Update(itemsLoadedMsg{items: sampleItems})
	return m2.(model)
}

func TestViewRendersDetail(t *testing.T) {
	m := loadedModel(t)
	m.table.SetCursor(1)
	m.width = 80
	view := m.View()
	if !strings.Contains(view, "node.exe") {
		t.Fatalf("view should contain selected process name:\n%s", view)
	}
	if !strings.Contains(view, "命令行") {
		t.Fatalf("view should contain cmdline detail section:\n%s", view)
	}

	// 命令行不可得时展示占位提示
	m2 := initialModel()
	m2r, _ := m2.Update(itemsLoadedMsg{items: []portscan.Item{{LocalHost: "0.0.0.0", Port: "5432", PID: "", Process: "pg"}}})
	mm := m2r.(model)
	mm.width = 80
	if v := mm.View(); !strings.Contains(v, "未获取到完整命令行") {
		t.Fatalf("view should show placeholder for missing cmdline:\n%s", v)
	}
}

func TestQuitOnQ(t *testing.T) {
	m := initialModel()
	m2, cmd := m.Update(runes("q"))
	if cmd == nil {
		t.Fatal("expect quit cmd on q")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("cmd should be tea.Quit")
	}
	if _, ok := m2.(model); !ok {
		t.Fatal("model type unchanged")
	}
}

func TestLoadAndFilterFlow(t *testing.T) {
	m := loadedModel(t)
	if len(m.all) != 2 || len(m.filtered) != 2 {
		t.Fatalf("load failed: %+v", m.all)
	}

	// "/" 进入过滤模式
	m2, _ := m.Update(runes("/"))
	m = m2.(model)
	if m.mode != modeFilter {
		t.Fatalf("mode = %v, want modeFilter", m.mode)
	}

	// 输入 80 实时过滤
	m2, _ = m.Update(runes("80"))
	m = m2.(model)
	if m.input.Value() != "80" {
		t.Fatalf("input = %q", m.input.Value())
	}
	if len(m.filtered) != 1 || m.filtered[0].Port != "8080" {
		t.Fatalf("filtered = %+v", m.filtered)
	}

	// enter 返回列表模式
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = m2.(model)
	if m.mode != modeList {
		t.Fatalf("mode = %v, want modeList", m.mode)
	}
}

func TestKillConfirmFlow(t *testing.T) {
	m := loadedModel(t)

	// 选中第一行后按 d，进入确认模式
	m.table.SetCursor(1)
	m2, _ := m.Update(runes("d"))
	m = m2.(model)
	if m.mode != modeConfirm || m.target.PID != "12345" {
		t.Fatalf("mode=%v target=%+v", m.mode, m.target)
	}

	// y 触发杀进程命令并回到列表模式（不实际执行 cmd）
	m2, cmd := m.Update(runes("y"))
	m = m2.(model)
	if m.mode != modeList || cmd == nil {
		t.Fatalf("mode=%v cmd=%v", m.mode, cmd)
	}

	// n 取消
	m2, _ = m.Update(runes("d"))
	m = m2.(model)
	m2, _ = m.Update(runes("n"))
	m = m2.(model)
	if m.mode != modeList {
		t.Fatalf("n should cancel, mode=%v", m.mode)
	}
}

func TestKillErrorShown(t *testing.T) {
	m := loadedModel(t)
	m2, _ := m.Update(killedMsg{target: sampleItems[1], err: errors.New("Access is denied")})
	m = m2.(model)
	if m.errMsg == "" || m.status != "" {
		t.Fatalf("errMsg=%q status=%q", m.errMsg, m.status)
	}
	if got := permHint(errors.New("Access is denied")); got == "" {
		t.Fatal("permHint should hit for Access is denied")
	}
}

func TestKillNoPID(t *testing.T) {
	m := initialModel()
	m2, _ := m.Update(itemsLoadedMsg{items: []portscan.Item{{LocalHost: "0.0.0.0", Port: "5432"}}})
	m = m2.(model)
	m2, _ = m.Update(runes("d"))
	m = m2.(model)
	if m.mode != modeList || m.errMsg == "" {
		t.Fatalf("no-PID row should not enter confirm, mode=%v errMsg=%q", m.mode, m.errMsg)
	}
}
