package network

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"
	"indigo6a.online/gokit/internal/portscan"
)

// PortItem 是共享扫描内核的记录类型别名
type PortItem = portscan.Item

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
		items, err := portscan.List(filterEntry.Text)
		if err != nil {
			handle.Hide()
			dialog.ShowError(err, w)
			return
		}
		portList = items
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
			o.(*widget.Label).SetText(portList[i.Row-1].LocalHost)
			o.Resize(fyne.NewSize(120, o.MinSize().Height))
		case 1:
			o.(*widget.Label).SetText(portList[i.Row-1].Port)
		case 2:
			o.(*widget.Label).SetText(portList[i.Row-1].PID)
		case 3:
			o.(*widget.Label).SetText(portList[i.Row-1].Process)
			o.Resize(fyne.NewSize(220, o.MinSize().Height))
		}
	})

	list.SetColumnWidth(0, 120)
	list.SetColumnWidth(1, 60)
	list.SetColumnWidth(2, 60)
	list.SetColumnWidth(3, 220)
	list.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}
		item := portList[id.Row-1]
		if id.Col == 3 {
			dialog.ShowConfirm("详情", "复制"+item.Process+"?", func(b bool) {
				clipboard.WriteAll(item.Process)
			}, w)
			return
		}
		if id.Col != 2 {
			return
		}
		dialog.ShowConfirm("提示", fmt.Sprintf("结束进程%v(%v)?", item.PID, item.Process), func(b bool) {
			if !b {
				return
			}
			pid, err := strconv.Atoi(item.PID)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if err := portscan.Kill(pid); err != nil {
				dialog.ShowError(err, w)
			}
		}, w)
	}
	return container.NewBorder(container.NewBorder(nil, nil, queryBtn, nil, filterEntry), nil, nil, nil, list)
}
