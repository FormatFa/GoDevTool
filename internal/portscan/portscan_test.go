package portscan

import (
	"reflect"
	"testing"
)

func TestSplitHostPort(t *testing.T) {
	cases := []struct {
		in, host, port string
	}{
		{"0.0.0.0:135", "0.0.0.0", "135"},
		{"[::]:135", "[::]", "135"},
		{":::8080", "::", "8080"},
		{"127.0.0.1:5000", "127.0.0.1", "5000"},
		{"localhost", "localhost", ""},
	}
	for _, c := range cases {
		host, port := splitHostPort(c.in)
		if host != c.host || port != c.port {
			t.Errorf("splitHostPort(%q) = (%q, %q), want (%q, %q)", c.in, host, port, c.host, c.port)
		}
	}
}

const winTasklistCSV = `"System","4","Services","0","24 K"
"node.exe","12345","Console","1","48,096 K"
"Windows Explorer","999","Console","1","80,120 K"
`

const winNetstat = `  协议  本地地址          外部地址        状态           PID

  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1044
  TCP    127.0.0.1:8084         127.0.0.1:54321        ESTABLISHED     12345
  TCP    [::]:135               [::]:0                 LISTENING       1044
  TCP    0.0.0.0:8084           0.0.0.0:0              LISTENING       12345
`

const winCmdlineCSV = `"ProcessId","CommandLine"
"1044",""
"12345","""C:\Program Files\nodejs\node.exe"" D:\work\server\app.js --port 8084 --中文参数"
`

func TestParseCmdlineCSV(t *testing.T) {
	got := parseCmdlineCSV(winCmdlineCSV)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got["1044"] != "" {
		t.Errorf("1044 = %q, want empty", got["1044"])
	}
	want := `"C:\Program Files\nodejs\node.exe" D:\work\server\app.js --port 8084 --中文参数`
	if got["12345"] != want {
		t.Errorf("12345 = %q, want %q", got["12345"], want)
	}
}

func TestParseWindows(t *testing.T) {
	cmdlines := parseCmdlineCSV(winCmdlineCSV)
	got := parseWindows(winTasklistCSV, winNetstat, cmdlines, "")
	want := []Item{
		{LocalHost: "0.0.0.0", Port: "135", PID: "1044", Process: ""},
		{LocalHost: "[::]", Port: "135", PID: "1044", Process: ""},
		{LocalHost: "0.0.0.0", Port: "8084", PID: "12345", Process: "node.exe",
			CommandLine: `"C:\Program Files\nodejs\node.exe" D:\work\server\app.js --port 8084 --中文参数`},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseWindows() = %+v, want %+v", got, want)
	}

	got = parseWindows(winTasklistCSV, winNetstat, cmdlines, "8084")
	if len(got) != 1 || got[0].Process != "node.exe" {
		t.Errorf("parseWindows(filter=8084) = %+v", got)
	}
}

const macPs = `  PID COMMAND
    1 /sbin/launchd
 1234 ssh -D 8080
  630 /usr/libexec/rapportd
`

const macLsof = `COMMAND   PID  USER   FD   TYPE   DEVICE SIZE/OFF NODE NAME
rapportd  630   root    4u  IPv4  0xabc      0t0  TCP *:54210 (LISTEN)
ControlCe 819   user   21u  IPv4  0xabd      0t0  TCP 127.0.0.1:5000 (LISTEN)
ssh      1234   user    5u  IPv6  0xabe      0t0  TCP [::1]:2222 (LISTEN)
launchd    1   root   10u  IPv6  0xabf      0t0  TCP *:22 (LISTEN)
`

func TestParseDarwin(t *testing.T) {
	got := parseDarwin(macPs, macLsof, "")
	want := []Item{
		{LocalHost: "*", Port: "22", PID: "1", Process: "launchd", CommandLine: "/sbin/launchd"},
		{LocalHost: "[::1]", Port: "2222", PID: "1234", Process: "ssh", CommandLine: "ssh -D 8080"},
		{LocalHost: "127.0.0.1", Port: "5000", PID: "819", Process: "ControlCe", CommandLine: ""},
		{LocalHost: "*", Port: "54210", PID: "630", Process: "rapportd", CommandLine: "/usr/libexec/rapportd"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseDarwin() = %+v, want %+v", got, want)
	}

	got = parseDarwin(macPs, macLsof, "2222")
	if len(got) != 1 || got[0].LocalHost != "[::1]" {
		t.Errorf("parseDarwin(filter=2222) = %+v", got)
	}
}

const linuxNetstat = `Active Internet connections (only servers)
Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN      987/sshd: /usr/sbin/sshd
tcp        0      0 127.0.0.1:6010          0.0.0.0:*               LISTEN      1234/sshd: user@pts/0
tcp6       0      0 :::8080                 :::*                    LISTEN      5555/docker-proxy
tcp        0      0 0.0.0.0:5432            0.0.0.0:*               LISTEN      -
udp        0      0 0.0.0.0:68              0.0.0.0:*                           -
`

func TestParseLinux(t *testing.T) {
	got := parseLinux(linuxNetstat, "")
	want := []Item{
		{LocalHost: "0.0.0.0", Port: "22", PID: "987", Process: "sshd: /usr/sbin/sshd"},
		{LocalHost: "0.0.0.0", Port: "5432", PID: "", Process: ""},
		{LocalHost: "127.0.0.1", Port: "6010", PID: "1234", Process: "sshd: user@pts/0"},
		{LocalHost: "::", Port: "8080", PID: "5555", Process: "docker-proxy"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseLinux() = %+v, want %+v", got, want)
	}

	got = parseLinux(linuxNetstat, "80")
	if len(got) != 1 || got[0].PID != "5555" {
		t.Errorf("parseLinux(filter=80) = %+v", got)
	}
}
