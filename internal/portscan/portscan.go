// Package portscan 提供跨平台的监听端口查询与进程终止能力，无 UI 依赖，供 GUI 与 TUI 共用。
package portscan

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// Item 表示一条监听端口记录
type Item struct {
	LocalHost   string
	Port        string
	PID         string
	Process     string
	CommandLine string
}

// List 获取本机监听端口列表。filterPort 为空返回全部，否则只返回端口包含该子串的记录。
func List(filterPort string) ([]Item, error) {
	switch runtime.GOOS {
	case "darwin":
		return listDarwin(filterPort)
	case "linux":
		return listLinux(filterPort)
	case "windows":
		return listWindows(filterPort)
	default:
		return nil, fmt.Errorf("portscan: unsupported platform %s", runtime.GOOS)
	}
}

// Kill 终止指定 PID 的进程（Windows 为 TerminateProcess，类 Unix 为 SIGKILL）。
func Kill(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("portscan: find process %d: %w", pid, err)
	}
	return proc.Kill()
}

var spaceRe = regexp.MustCompile(`\s+`)

// splitHostPort 按最后一个冒号拆分地址，兼容 IPv6（如 [::]:135、:::8080）。
func splitHostPort(s string) (host, port string) {
	i := strings.LastIndex(s, ":")
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i+1:]
}

func sortByPort(items []Item) {
	sort.SliceStable(items, func(i, j int) bool {
		v1, _ := strconv.Atoi(items[i].Port)
		v2, _ := strconv.Atoi(items[j].Port)
		return v1 < v2
	})
}

// firstToken 取命令行的第一个词（可执行文件路径）
func firstToken(s string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

// baseName 取路径的最后一段作为进程名
func baseName(s string) string {
	if i := strings.LastIndexAny(s, `/\`); i >= 0 {
		return s[i+1:]
	}
	return s
}

// ---------------- Windows ----------------

func listWindows(filterPort string) ([]Item, error) {
	// CSV 格式可正确处理含空格/逗号的进程名
	tasklistOut, err := exec.Command("tasklist", "/FO", "CSV", "/NH").Output()
	if err != nil {
		return nil, fmt.Errorf("portscan: exec tasklist: %w", err)
	}
	netstatOut, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return nil, fmt.Errorf("portscan: exec netstat: %w", err)
	}
	// 完整命令行获取失败时降级为空，不阻塞端口列表
	cmdlines, _ := windowsCmdlines()
	return parseWindows(string(tasklistOut), string(netstatOut), cmdlines, filterPort), nil
}

// windowsCmdlines 通过 PowerShell CIM 查询全量进程命令行（UTF-8 CSV）
func windowsCmdlines() (map[string]string, error) {
	const script = "[Console]::OutputEncoding=[Text.Encoding]::UTF8; " +
		"Get-CimInstance Win32_Process | Select-Object ProcessId,CommandLine | ConvertTo-Csv -NoTypeInformation"
	out, err := exec.Command("powershell", "-NoProfile", "-Command", script).Output()
	if err != nil {
		return nil, fmt.Errorf("portscan: exec powershell Get-CimInstance: %w", err)
	}
	return parseCmdlineCSV(string(out)), nil
}

// parseCmdlineCSV 解析 ConvertTo-Csv 输出为 pid -> 命令行 映射
func parseCmdlineCSV(csvOut string) map[string]string {
	m := map[string]string{}
	r := csv.NewReader(strings.NewReader(csvOut))
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return m
	}
	for i, fields := range records {
		if i == 0 || len(fields) < 2 { // 跳过表头 "ProcessId","CommandLine"
			continue
		}
		m[fields[0]] = fields[1]
	}
	return m
}

func parseWindows(tasklistCSV, netstatOut string, cmdlines map[string]string, filterPort string) []Item {
	tasklist := map[string]string{}
	r := csv.NewReader(strings.NewReader(tasklistCSV))
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err == nil {
		for _, fields := range records {
			if len(fields) >= 2 {
				tasklist[fields[1]] = fields[0]
			}
		}
	}

	items := []Item{}
	for _, line := range strings.Split(netstatOut, "\n") {
		if !strings.Contains(line, "LISTENING") {
			continue
		}
		cols := spaceRe.Split(strings.TrimSpace(line), -1)
		// [ TCP 0.0.0.0:135 0.0.0.0:0 LISTENING 1044 ]
		if len(cols) < 5 {
			continue
		}
		host, port := splitHostPort(cols[1])
		if port == "" || !strings.Contains(port, filterPort) {
			continue
		}
		pid := cols[4]
		items = append(items, Item{
			LocalHost:   host,
			Port:        port,
			PID:         pid,
			Process:     tasklist[pid],
			CommandLine: cmdlines[pid],
		})
	}
	sortByPort(items)
	return items
}

// ---------------- macOS ----------------

func listDarwin(filterPort string) ([]Item, error) {
	// -o pid,command 保证命令列完整输出（管道下不截断）
	psOut, err := exec.Command("ps", "-axo", "pid,command").Output()
	if err != nil {
		return nil, fmt.Errorf("portscan: exec ps: %w", err)
	}
	// -n 禁止转换网络数字到域名，-P 禁止转换端口到端口名
	lsofOut, err := exec.Command("lsof", "-nP", "-i").Output()
	if err != nil {
		return nil, fmt.Errorf("portscan: exec lsof: %w", err)
	}
	return parseDarwin(string(psOut), string(lsofOut), filterPort), nil
}

func parseDarwin(psOut, lsofOut, filterPort string) []Item {
	// ps -axo pid,command: [ PID 完整命令（可含空格） ]
	cmds := map[string]string{}
	for _, line := range strings.Split(psOut, "\n") {
		cols := spaceRe.Split(strings.TrimSpace(line), -1)
		if len(cols) >= 2 {
			if _, err := strconv.Atoi(cols[0]); err == nil {
				cmds[cols[0]] = strings.Join(cols[1:], " ")
			}
		}
	}

	items := []Item{}
	for _, line := range strings.Split(lsofOut, "\n") {
		if !strings.Contains(line, "LISTEN") {
			continue
		}
		cols := spaceRe.Split(strings.TrimSpace(line), -1)
		// [ COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME ]，cols[8] 为本地地址
		if len(cols) < 9 {
			continue
		}
		host, port := splitHostPort(cols[8])
		if port == "" || !strings.Contains(port, filterPort) {
			continue
		}
		pid := cols[1]
		// 进程名：ps 完整命令首词取 basename，缺失时回退 lsof COMMAND（可能被截断）
		process := baseName(firstToken(cmds[pid]))
		if process == "" {
			process = cols[0]
		}
		items = append(items, Item{
			LocalHost:   host,
			Port:        port,
			PID:         pid,
			Process:     process,
			CommandLine: cmds[pid],
		})
	}
	sortByPort(items)
	return items
}

// ---------------- Linux ----------------

func listLinux(filterPort string) ([]Item, error) {
	netstatOut, err := exec.Command("netstat", "-tlnp").Output()
	if err != nil {
		return nil, fmt.Errorf("portscan: exec netstat: %w", err)
	}
	return parseLinux(string(netstatOut), filterPort), nil
}

func parseLinux(netstatOut, filterPort string) []Item {
	items := []Item{}
	for _, line := range strings.Split(netstatOut, "\n") {
		if !strings.Contains(line, "LISTEN") {
			continue
		}
		cols := spaceRe.Split(strings.TrimSpace(line), -1)
		// [ tcp 0 0 0.0.0.0:22 0.0.0.0:* LISTEN 987/sshd: /usr/sbin/sshd ]
		if len(cols) < 6 {
			continue
		}
		host, port := splitHostPort(cols[3])
		if port == "" || !strings.Contains(port, filterPort) {
			continue
		}
		pid, process := "", ""
		if len(cols) >= 7 && cols[6] != "-" {
			// PID/Program name 可能含空格（如 "987/sshd: /usr/sbin/sshd"），先拼回整列再拆
			full := strings.Join(cols[6:], " ")
			if i := strings.Index(full, "/"); i >= 0 {
				if _, err := strconv.Atoi(strings.TrimSpace(full[:i])); err == nil {
					pid = strings.TrimSpace(full[:i])
					process = full[i+1:]
				}
			}
		}
		item := Item{LocalHost: host, Port: port, PID: pid, Process: process}
		// 完整命令行：优先读 /proc（内核线程与 hidepid 下为空，保持 netstat 结果）
		if pid != "" {
			if b, err := os.ReadFile("/proc/" + pid + "/cmdline"); err == nil && len(b) > 0 {
				args := strings.Split(strings.TrimRight(string(b), "\x00"), "\x00")
				item.CommandLine = strings.Join(args, " ")
				if name := baseName(firstToken(item.CommandLine)); name != "" {
					item.Process = name
				}
			}
		}
		items = append(items, item)
	}
	sortByPort(items)
	return items
}
