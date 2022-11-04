package network

import (
	"log"
	"os"
	"os/exec"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fyne.io/fyne/v2"
)

func EditHost(w fyne.Window) fyne.CanvasObject {
	return windows(w)
}
func windows(w fyne.Window) fyne.CanvasObject {
	editBtn := widget.NewButton("编辑", func() {
		log.Println(os.TempDir())
		tempScriptPath := os.TempDir() + "\\godev-host.bat"
		file, err := os.Create(tempScriptPath)
		if err != nil {
			return
		}
		script := `@echo off
		%1 %2
		mshta vbscript:createobject("shell.application").shellexecute("%~s0","goto :runas","","runas",1)(window.close)&goto :uacfalse
		:runas
		echo uca success
		notepad C:\Windows\System32\drivers\etc\hosts
		goto :eof
		
		:uacfalse
		echo get uca fail
		pause
		 
		`
		file.WriteString(script)
		file.Close()
		// script = "echo hello"
		cmd := exec.Command("cmd.exe", "/C", tempScriptPath)
		out, cmderr := cmd.CombinedOutput()
		if cmderr != nil {
			log.Fatalf("cmd fail:5%s\n", cmderr)
		} else {
			log.Println(string(out))
		}

	})

	label := widget.NewLabel("C:\\Windows\\System32\\drivers\\etc\\hosts")
	return container.New(layout.NewVBoxLayout(), label, editBtn)

}
