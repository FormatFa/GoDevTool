package network

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func IpTool(w fyne.Window) fyne.CanvasObject {
	return windowsIpTool(w)
}
func windowsIpTool(w fyne.Window) fyne.CanvasObject {

	publicIp := widget.NewMultiLineEntry()

	// https://github.com/ihmily/ip-info-api#address-1.1
	api := "http://ip-api.com/json/"
	apiSelect := widget.NewSelect([]string{"http://ip-api.com/json/", "https://webapi-pc.meitu.com/common/ip_location", "https://ip.useragentinfo.com/json"}, func(s string) {
		api = s
		fmt.Println(api)
	})
	getPublicIpBtn := widget.NewButton("获取", func() {
		response, err := http.Get(api)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err == nil {
			strRes := string(body)
			publicIp.Text = publicIp.Text + "\n" + strRes
			fmt.Println(strRes)
		}

	})

	myIp := widget.NewMultiLineEntry()
	getMyIp := widget.NewButton("获取", func() {
		interfaces, err := net.Interfaces()

		if err != nil {

			fmt.Println("Error:", err)

			return

		}

		var sb strings.Builder

		for _, iface := range interfaces {
			// 排除一些特殊接口
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			// 获取接口的地址信息
			addrs, err := iface.Addrs()
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			// 遍历接口的地址
			for _, addr := range addrs {
				// 检查地址类型
				switch v := addr.(type) {
				case *net.IPNet:
					// IPv4 或 IPv6 地址
					fmt.Println(v.IP.String())
					sb.WriteString(iface.Name + " " + v.IP.String() + "\n")
				case *net.IPAddr:
					// 一般情况下是 IPv4 地址
					fmt.Println(v.IP)
					sb.WriteString(iface.Name + " " + v.IP.String() + "\n")
				}

			}

		}
		fmt.Println("ff:" + sb.String())
		myIp.Text = sb.String()
	})

	return container.NewVBox(
		widget.NewLabel("公网IP"),
		container.NewHBox(apiSelect, getPublicIpBtn),
		publicIp,
		widget.NewLabel("本机ip"),
		getMyIp,
		myIp,
	)

}
