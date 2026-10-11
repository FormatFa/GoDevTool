package mirror

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"indigo6a.online/gokit/pkg/mirror"
)

// PipConfigView creates and returns the pip configuration view
func PipConfigView(w fyne.Window) fyne.CanvasObject {
	// Create pip manager
	pipMgr := mirror.NewPipManager()

	// Detect current configuration
	configInfo := pipMgr.Detect()

	// UI elements
	pathLabel := widget.NewLabel("配置文件: " + configInfo.ConfigFile)

	statusText := "当前状态: " + string(configInfo.Status)
	if configInfo.Status == mirror.StatusMirror && configInfo.CurrentURL != "" {
		statusText += "\n当前镜像: " + configInfo.CurrentURL
	}
	statusLabel := widget.NewLabel(statusText)

	// Mirror selection
	mirrorNames := mirror.GetMirrorNames("pip")
	mirrorSelect := widget.NewSelect(mirrorNames, nil)
	if len(mirrorNames) > 0 {
		mirrorSelect.SetSelected(mirrorNames[0])
	}

	// Configure button
	configBtn := widget.NewButton("配置镜像源", func() {
		selectedMirror := mirrorSelect.Selected
		if selectedMirror == "" {
			dialog.ShowInformation("提示", "请选择镜像源", w)
			return
		}

		mirrorConfig, ok := mirror.GetMirrorByName("pip", selectedMirror)
		if !ok {
			dialog.ShowInformation("错误", "找不到镜像配置", w)
			return
		}

		dialog.ShowConfirm("确认配置",
			"确定要配置 pip 使用 "+selectedMirror+" 镜像源吗？\n\n配置文件: "+configInfo.ConfigFile,
			func(confirmed bool) {
				if confirmed {
					if err := pipMgr.Configure(mirrorConfig); err != nil {
						dialog.ShowInformation("配置失败", err.Error(), w)
						log.Println("Pip config error:", err)
						return
					}
					dialog.ShowInformation("配置成功", "pip 已配置使用 "+selectedMirror+" 镜像源", w)
					// Refresh view
					w.SetContent(PipConfigView(w))
				}
			}, w)
	})

	// Restore button
	restoreBtn := widget.NewButton("恢复默认", func() {
		dialog.ShowConfirm("确认恢复",
			"确定要恢复 pip 到默认配置吗？",
			func(confirmed bool) {
				if confirmed {
					if err := pipMgr.Restore(); err != nil {
						dialog.ShowInformation("恢复失败", err.Error(), w)
						log.Println("Pip restore error:", err)
						return
					}
					dialog.ShowInformation("恢复成功", "pip 已恢复到默认配置", w)
					// Refresh view
					w.SetContent(PipConfigView(w))
				}
			}, w)
	})

	// Layout
	return container.NewVBox(
		widget.NewLabel("pip 镜像源配置"),
		widget.NewSeparator(),
		pathLabel,
		widget.NewLabel(""),
		statusLabel,
		widget.NewLabel(""),
		widget.NewLabel("选择镜像源:"),
		mirrorSelect,
		widget.NewLabel(""),
		container.NewHBox(configBtn, restoreBtn),
	)
}
