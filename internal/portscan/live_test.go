package portscan

import (
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"testing"
	"time"
)

// TestLiveList 在真实系统上执行扫描，需设置 GO_KIT_LIVE_TEST=1 才运行
func TestLiveList(t *testing.T) {
	if os.Getenv("GO_KIT_LIVE_TEST") == "" {
		t.Skip("set GO_KIT_LIVE_TEST=1 to run live scan")
	}
	items, err := List("")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("no listening ports found")
	}
	for _, it := range items[:min(5, len(items))] {
		t.Logf("%+v", it)
	}
}

// TestLiveKill 端到端验证：起一个监听 18080 的子进程 → List 定位 → Kill → 确认退出
func TestLiveKill(t *testing.T) {
	if os.Getenv("GO_KIT_LIVE_TEST") == "" {
		t.Skip("set GO_KIT_LIVE_TEST=1 to run live kill")
	}
	if runtime.GOOS != "windows" {
		t.Skip("windows-only helper")
	}

	// 先预留一个随机空闲端口，关闭后交给子进程使用（避免撞上已占用端口）
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	_ = ln.Close()

	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`$l=[System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback,`+port+`);$l.Start();Start-Sleep -Seconds 120`)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill() }()
	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()

	deadline := time.Now().Add(15 * time.Second)
	var found *Item
	for time.Now().Before(deadline) {
		items, err := List(port)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) > 0 {
			found = &items[0]
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	if found == nil {
		t.Fatal("port " + port + " not found in scan results")
	}
	if pid, _ := strconv.Atoi(found.PID); pid != cmd.Process.Pid {
		t.Fatalf("scanned pid %s != child pid %d", found.PID, cmd.Process.Pid)
	}

	pid, _ := strconv.Atoi(found.PID)
	if err := Kill(pid); err != nil {
		t.Fatalf("Kill: %v", err)
	}
	select {
	case <-waitCh:
		t.Logf("child process (pid %d, port %s) terminated", pid, found.Port)
	case <-time.After(5 * time.Second):
		t.Fatal("child process did not exit after Kill")
	}
}
