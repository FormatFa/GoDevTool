// gokit 是 GoDevKit 的 TUI 命令行版：端口占用查询与进程查杀。
//
// 按键: ↑/↓ 移动, d 杀进程, / 过滤, r 刷新, q 退出
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"indigo6a.online/gokit/internal/portscan"
)

// ----- 异步消息 -----

type itemsLoadedMsg struct {
	items []portscan.Item
	err   error
}

type killedMsg struct {
	target portscan.Item
	err    error
}

func loadCmd() tea.Cmd {
	return func() tea.Msg {
		items, err := portscan.List("")
		return itemsLoadedMsg{items: items, err: err}
	}
}

func killCmd(target portscan.Item) tea.Cmd {
	return func() tea.Msg {
		pid, err := strconv.Atoi(target.PID)
		if err != nil {
			return killedMsg{target: target, err: err}
		}
		return killedMsg{target: target, err: portscan.Kill(pid)}
	}
}

// ----- 快捷键 -----

type keyMap struct {
	up      key.Binding
	filter  key.Binding
	refresh key.Binding
	kill    key.Binding
	quit    key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/↓", "移动")),
		filter:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "过滤")),
		refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "刷新")),
		kill:    key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "杀进程")),
		quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "退出")),
	}
}

// ShortHelp 实现 help.KeyMap
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.up, k.kill, k.filter, k.refresh, k.quit}
}

// FullHelp 实现 help.KeyMap
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.up}, {k.kill}, {k.filter}, {k.refresh}, {k.quit}}
}

// ----- 模型 -----

const (
	modeList = iota
	modeFilter
	modeConfirm
)

type model struct {
	table    table.Model
	input    textinput.Model
	help     help.Model
	keys     keyMap
	all      []portscan.Item
	filtered []portscan.Item
	mode     int
	target   portscan.Item
	loading  bool
	status   string
	errMsg   string
	width    int
}

func initialModel() model {
	columns := []table.Column{
		{Title: "本地地址", Width: 18},
		{Title: "端口", Width: 8},
		{Title: "PID", Width: 9},
		{Title: "进程", Width: 34},
	}
	t := table.New(table.WithColumns(columns), table.WithRows(nil),
		table.WithFocused(true), table.WithHeight(12), table.WithWidth(74))
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236"))
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
	t.SetStyles(s)

	input := textinput.New()
	input.Placeholder = "输入端口关键字，如 8080"
	input.Prompt = "过滤 > "

	return model{
		table:   t,
		input:   input,
		help:    help.New(),
		keys:    newKeyMap(),
		loading: true,
	}
}

// Init 启动时自动加载端口列表
func (m model) Init() tea.Cmd {
	return loadCmd()
}

// applyFilter 在全量数据上按端口做包含匹配，重建表格行
func (m *model) applyFilter() {
	f := m.input.Value()
	m.filtered = m.filtered[:0]
	for _, it := range m.all {
		if strings.Contains(it.Port, f) {
			m.filtered = append(m.filtered, it)
		}
	}
	rows := make([]table.Row, 0, len(m.filtered))
	for _, it := range m.filtered {
		host, pid, process := it.LocalHost, it.PID, it.Process
		if host == "" {
			host = "*"
		}
		if pid == "" {
			pid = "-"
		}
		if process == "" {
			process = "-"
		}
		rows = append(rows, table.Row{host, it.Port, pid, process})
	}
	m.table.SetRows(rows)
}

// permHint 对常见权限错误补充提示
func permHint(err error) string {
	s := strings.ToLower(err.Error())
	for _, kw := range []string{"access is denied", "拒绝访问", "operation not permitted", "permission denied"} {
		if strings.Contains(s, kw) {
			return "（可能需要管理员/超级用户权限运行）"
		}
	}
	return ""
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.table.SetWidth(min(msg.Width-2, 74))
		h := msg.Height - 14
		if h < 5 {
			h = 5
		}
		m.table.SetHeight(h)
		return m, nil

	case itemsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errMsg = "加载端口列表失败: " + msg.err.Error()
			return m, nil
		}
		m.errMsg, m.status = "", ""
		m.all = msg.items
		m.applyFilter()
		return m, nil

	case killedMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("结束进程 %s(%s) 失败: %v%s", msg.target.Process, msg.target.PID, msg.err, permHint(msg.err))
			m.status = ""
		} else {
			m.errMsg = ""
			m.status = fmt.Sprintf("已结束进程 %s(%s)，正在刷新…", msg.target.Process, msg.target.PID)
		}
		m.loading = true
		return m, loadCmd()

	case tea.KeyMsg:
		switch m.mode {
		case modeConfirm:
			switch msg.String() {
			case "y":
				target := m.target
				m.mode = modeList
				m.status = "正在结束进程 " + target.Process + "…"
				return m, killCmd(target)
			case "n", "esc":
				m.mode = modeList
			}
			return m, nil

		case modeFilter:
			switch msg.String() {
			case "enter", "esc":
				m.mode = modeList
				m.input.Blur()
				m.table.Focus()
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				m.applyFilter()
				return m, cmd
			}
			return m, nil

		default: // modeList
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "r":
				m.loading = true
				m.status = "刷新中…"
				return m, loadCmd()
			case "/":
				m.mode = modeFilter
				m.table.Blur()
				return m, m.input.Focus()
			case "d", "x":
				if len(m.filtered) == 0 {
					return m, nil
				}
				target := m.filtered[m.table.Cursor()]
				if target.PID == "" {
					m.status = ""
					m.errMsg = "该记录无 PID（可能权限不足未获取到），无法结束"
					return m, nil
				}
				m.target = target
				m.mode = modeConfirm
				return m, nil
			}
			// 其余按键（方向键、翻页等）交给表格
		}
	}

	switch m.mode {
	case modeFilter:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	default:
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}
}

// ----- 视图 -----

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Padding(0, 1)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Padding(0, 1)
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Padding(0, 1)
	confirmStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Padding(0, 1)
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1)
	detailStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

func (m model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("GoDevKit TUI — 端口占用查询 / 杀进程") + "\n\n")

	if m.loading && len(m.all) == 0 {
		b.WriteString("  加载中…\n")
	} else if len(m.filtered) == 0 {
		b.WriteString("  无匹配结果（/ 修改过滤，r 刷新）\n")
	} else {
		b.WriteString(m.table.View() + "\n")

		// 详情区：选中行的完整命令行
		it := m.filtered[min(m.table.Cursor(), len(m.filtered)-1)]
		cl := it.CommandLine
		if cl == "" {
			cl = "（未获取到完整命令行，可能为系统进程或权限不足）"
		}
		name := it.Process
		if name == "" {
			name = "-"
		}
		detail := detailStyle.
			Width(max(m.width-4, 20)).
			MaxHeight(3).
			Render(fmt.Sprintf("命令行 [%s · PID %s]: %s", name, it.PID, cl))
		b.WriteString("\n" + detail + "\n")
	}

	b.WriteString("\n  " + m.input.View() + "\n")

	if m.errMsg != "" {
		b.WriteString("\n" + errorStyle.Render(m.errMsg))
	} else if m.status != "" {
		b.WriteString("\n" + statusStyle.Render(m.status))
	}

	if m.mode == modeConfirm {
		name := m.target.Process
		if name == "" {
			name = "?"
		}
		b.WriteString("\n" + confirmStyle.Render(fmt.Sprintf("结束进程 %s (PID %s)？ [y/n]", name, m.target.PID)))
	}

	var helpLine string
	switch m.mode {
	case modeFilter:
		helpLine = "enter 应用过滤 • esc 返回"
	case modeConfirm:
		helpLine = "y 确认 • n 取消"
	default:
		helpLine = m.help.View(m.keys)
	}
	b.WriteString("\n\n" + helpStyle.Render(helpLine))
	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("运行出错:", err)
		os.Exit(1)
	}
}
