package network

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"
)

type PortItem struct {
	localHost string
	port      string
	pid       string
	process   string
}

func PortUseView(w fyne.Window) fyne.CanvasObject {
	var list *widget.Table
	portList := []PortItem{}
	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("输入端口过滤")
	filterEntry.OnChanged = func(s string) {
	}
	queryBtn := widget.NewButton("查询", func() {
		handle := dialog.NewCustom("提示", "加载中", widget.NewProgressBarInfinite(), w)
		handle.Show()
		portList = getPortList(filterEntry.Text)
		list.Refresh()
		handle.Hide()
	})

	header := []string{"localHost", "Port", "Pid", "Note"}

	list = widget.NewTable(func() (int, int) {
		// add header
		return len(portList) + 1, 4
	}, func() fyne.CanvasObject {
		text := widget.NewLabel("8084")
		text.Wrapping = fyne.TextWrapOff
		return text
	}, func(i widget.TableCellID, o fyne.CanvasObject) {
		if i.Row == 0 {
			o.(*widget.Label).SetText(header[i.Col])
			return
		}
		switch i.Col {
		case 0:
			o.(*widget.Label).SetText(portList[i.Row-1].localHost)
			o.Resize(fyne.NewSize(120, o.MinSize().Height))
		case 1:
			o.(*widget.Label).SetText(portList[i.Row-1].port)
		case 2:
			o.(*widget.Label).SetText(portList[i.Row-1].pid)
		case 3:
			o.(*widget.Label).SetText(portList[i.Row-1].process)
			o.Resize(fyne.NewSize(220, o.MinSize().Height))
		}
	})

	list.SetColumnWidth(0, 120)
	list.SetColumnWidth(1, 60)
	list.SetColumnWidth(2, 60)
	list.SetColumnWidth(3, 220)
	list.OnSelected = func(id widget.TableCellID) {
		if id.Col == 3 {
			dialog.ShowConfirm("详情", "复制"+portList[id.Row-1].process+"?", func(b bool) {
				clipboard.WriteAll(portList[id.Row-1].process)
			}, w)
			return
		}
		if id.Col != 2 {
			return
		}
		dialog.ShowConfirm("提示", fmt.Sprintf("结束进程%v(%v)?", portList[id.Row].pid, portList[id.Row].process), func(b bool) {
			if b {
				pid, err := strconv.Atoi(portList[id.Row].pid)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				pro, err := os.FindProcess(pid)
				if err != nil {
					if err != nil {
						dialog.ShowError(err, w)
						return
					}
				}
				err = pro.Kill()
				if err != nil {
					if err != nil {
						dialog.ShowError(err, w)
						return
					}
				}
			}
		}, w)
	}
	return container.NewBorder(container.NewBorder(nil, nil, queryBtn, nil, filterEntry), nil, nil, nil, list)
}

// windows darwin linux android
func getPortList(filterPort string) []PortItem {
	if runtime.GOOS == "darwin" {
		return getDarwinPortList(filterPort)
	} else {
		return getWindowsPortList(filterPort)

	}
}
func getDarwinPortList(filterPort string) []PortItem {
	portList := []PortItem{}
	// netstat -ano|findstr LISTENING
	spaceRe, _ := regexp.Compile(`\s+`)

	tasklist := map[string]string{}
	log.Println("exec: ps -axl")
	out, err := exec.Command("ps", "-axl").Output()
	if err != nil {
		log.Fatal(err)
	}
	for _, temp := range strings.Split(string(out), "\n") {
		cols := spaceRe.Split(temp, 16)
		if len(cols) > 14 {
			tasklist[cols[2]] = cols[15]
		}
		// fmt.Println("tabs:", strings.Split(temp, "                    "))
	}
	log.Println("exec: lsof -nP -i")
	// 获取端口列表 -n 禁止转换网络数字到域名，-P禁止转换端口到端口名
	out, err = exec.Command("lsof", "-nP", "-i").Output()
	if err != nil {
		log.Fatal(err)
	}
	result := string(out)
	for _, temp := range strings.Split(result, "\n") {
		if strings.Contains(temp, "LISTEN") {
			cols := spaceRe.Split(temp, -1)

			// [ TCP 0.0.0.0:135 0.0.0.0:0 LISTENING 1044 ]
			if len(cols) >= 6 {
				var host, port string = "", ""
				splited := strings.Split(cols[8], ":")
				if len(splited) == 2 {
					host = splited[0]
					port = splited[1]
				} else {
					port = splited[1]
				}
				if strings.Contains(port, filterPort) && port != "" {
					portList = append(portList, PortItem{localHost: host, port: port, pid: cols[1], process: tasklist[cols[1]]})
				}
			}
		}

	}
	sort.SliceStable(portList, func(i, j int) bool {
		v1, _ := strconv.Atoi(portList[i].port)
		v2, _ := strconv.Atoi(portList[j].port)
		return v1 < v2
	})
	return portList
}
func getLinuxPortList(filterPort string) []PortItem {
	return nil
}
func getWindowsPortList(filterPort string) []PortItem {
	portList := []PortItem{}
	// netstat -ano|findstr LISTENING
	spaceRe, _ := regexp.Compile(`\s+`)

	tasklist := map[string]string{}
	log.Println("exec: tasklist")
	out, err := exec.Command("tasklist").Output()
	if err != nil {
		log.Fatal(err)
	}
	for _, temp := range strings.Split(string(out), "\n") {
		cols := spaceRe.Split(temp, -1)
		if len(cols) > 2 {
			tasklist[cols[1]] = cols[0]
		}
		// fmt.Println("tabs:", strings.Split(temp, "                    "))
	}
	log.Println("exec: netstat -ano")
	// 获取端口列表
	out, err = exec.Command("netstat", "-ano").Output()
	if err != nil {
		log.Fatal(err)
	}
	result := string(out)
	for _, temp := range strings.Split(result, "\n") {
		if strings.Contains(temp, "LISTENING") {
			cols := spaceRe.Split(temp, -1)

			// [ TCP 0.0.0.0:135 0.0.0.0:0 LISTENING 1044 ]
			if len(cols) >= 6 {
				var host, port string = "", ""
				splited := strings.Split(cols[2], ":")
				if len(splited) == 2 {
					host = splited[0]
					port = splited[1]
				} else {
					port = splited[1]
				}
				if strings.Contains(port, filterPort) && port != "" {
					portList = append(portList, PortItem{localHost: host, port: port, pid: cols[5], process: tasklist[cols[5]]})
				}
			}
		}

	}
	sort.SliceStable(portList, func(i, j int) bool {
		v1, _ := strconv.Atoi(portList[i].port)
		v2, _ := strconv.Atoi(portList[j].port)
		return v1 < v2
	})
	return portList
}
