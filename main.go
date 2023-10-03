package main

import (
	"image/color"
	"log"
	"os"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
	"indigo6a.online/gokit/views/encode"
	"indigo6a.online/gokit/views/network"
)

func init() {
	//设置中文环境
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		// fmt.Println(path)
		//楷体:simkai.ttf
		//黑体:simhei.ttf
		if strings.Contains(path, "simhei.ttf") {
			log.Println("设置字体路径=", path)
			os.Setenv("FYNE_FONT", path)
			break
		}
		//mac
		if strings.Contains(path, "Arial Unicode.ttf") {
			log.Println("设置字体路径=", path)
			os.Setenv("FYNE_FONT", path)
			break
		}
		//mac
		if strings.Contains(path, "Apple Braille.ttf") {
			log.Println("设置字体路径=", path)
			os.Setenv("FYNE_FONT", path)
			break
		}

	}
}

func main() {
	myApp := app.New()
	settings := myApp.Settings()

	// 设置主题为黑夜主题，固定住是因为dns查询里有设置字体颜色，mac超级用户运行时会是白色主题，白色主题下会显示有问题
	settings.SetTheme(theme.DarkTheme())

	w := myApp.NewWindow("GoDevKit")

	helpMenu := fyne.NewMenu("帮助", fyne.NewMenuItem("关于", func() {
		dialog.ShowInformation("关于", "v0.0.1.221014_alpha", w)
	}), fyne.NewMenuItem("更新日志", func() {
		dialog.ShowInformation("更新日志", "about..", w)
	}))
	w.SetMainMenu(fyne.NewMainMenu(helpMenu))

	menu := map[string][]string{"": {"网络工具", "加密工具"}, "加密工具": {"常用加密"}, "网络工具": {"端口检测", "host文件", "DNS查询"}, "后端开发": {"测试1"}, "前端开发": {"工具1"}, "移动开发": {"Apk文件解析"}}
	left := widget.NewTreeWithStrings(menu)
	left.Resize(fyne.NewSize(90, left.MinSize().Height))

	// 初始化
	right := canvas.NewRectangle(color.Black)

	content := container.New(layout.NewBorderLayout(nil, nil, left, right), left, layout.NewSpacer(), right)
	content.Resize(fyne.NewSize(500, 500))
	w.SetContent(content)

	w.Resize(fyne.NewSize(500, 500))

	// 点击切换选项卡
	left.OnSelected = func(uid widget.TreeNodeID) {
		if uid == "常用加密" {
			content.Objects[1] = encode.Base64View(w)
		} else if uid == "端口检测" {
			content.Objects[1] = network.PortUseView(w)
		} else if uid == "host文件" {
			content.Objects[1] = network.EditHost(w)
		} else if uid == "DNS查询" {
			content.Objects[1] = network.DnsTool(w)
		}

		content.Objects[1].Resize(fyne.NewSize(400, right.Size().Height))

		content.Refresh()
	}
	log.Println("OS=", runtime.GOOS)
	w.ShowAndRun()
}
