package network

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fyne.io/fyne/v2"
)

// 暂时在mac上测试通过，需要超级权限运行

func DnsTool(w fyne.Window) fyne.CanvasObject {
	return windowsDns(w)
}

// TODO 设置不同的颜色，设置超时时间，端口可配置,多线程查询
func windowsDns(w fyne.Window) fyne.CanvasObject {

	urlText := ""
	ipList := []string{}
	dnsList := []string{}
	reqTimeList := []int{}
	urlInput := widget.NewMultiLineEntry()
	urlInput.SetPlaceHolder("需要连接的域名，逗号分隔")
	urlInput.SetText("https://dl.google.com")

	dnsInput := widget.NewMultiLineEntry()
	dnsInput.SetPlaceHolder("输入dns")
	dnsInput.SetText("1.1.1.1,114.114.114.114")

	inputLayout := container.New(layout.NewVBoxLayout(), urlInput, dnsInput)

	label := widget.NewLabel("C:\\Windows\\System32\\drivers\\etc\\hosts")

	table := widget.NewTable(func() (int, int) {
		return len(ipList), 4
	}, func() fyne.CanvasObject {
		return canvas.NewText("null", color.White)
	}, func(i widget.TableCellID, o fyne.CanvasObject) {
		text := o.(*canvas.Text)
		if i.Col == 0 {
			// url
			text.Text = urlText
		} else if i.Col == 1 {
			// dns
			text.Text = dnsList[i.Row]
		} else if i.Col == 2 {
			// ip
			text.Text = ipList[i.Row]

		} else if i.Col == 3 {
			time := reqTimeList[i.Row]
			if time == -1 {
				text.Text = "超时(" + strconv.Itoa(int(reqTimeList[i.Row])) + ")"
			} else {
				text.Text = strconv.Itoa(int(reqTimeList[i.Row])) + "ms"
			}
			if time <= 1000 && time > 0 {
				text.Color = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
			} else if time > 1000 && time <= 2000 {
				text.Color = color.NRGBA{R: 255, G: 255, B: 0, A: 255}
			} else {
				text.Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
			}

		}

	})
	table.SetColumnWidth(0, 220)
	table.SetColumnWidth(1, 220)
	table.SetColumnWidth(2, 110)
	table.SetColumnWidth(3, 120)

	editBtn := widget.NewButton("查询", func() {
		log.Println("text:" + urlInput.Text)

		urlText = urlInput.Text
		// must have schema or null
		if !strings.HasPrefix(urlText, "http") {
			urlText = "http://" + urlText
		}

		u, err := url.Parse(urlText)
		if err != nil {
			fmt.Println("解析网址时发生错误:", err)
			return
		}
		urlText = u.Host
		log.Println("url text:" + urlText)
		dnsList = strings.Split(dnsInput.Text, ",")
		log.Println(dnsList)

		ipList = []string{}
		reqTimeList = []int{}
		for _, value := range dnsList {
			log.Println("process dns :" + value)
			r := &net.Resolver{
				PreferGo: false,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{
						Timeout: 10 * time.Second,
					}
					return d.DialContext(ctx, "udp", value+":53")
				},
			}

			ips, _ := r.LookupHost(context.Background(), urlText)
			log.Println(ips)
			if len(ips) > 0 {
				nowIp := ips[0]
				ipList = append(ipList, nowIp)
				reqTimeList = append(reqTimeList, GetRequestTime(nowIp))
			} else {
				ipList = append(ipList, "null")
				reqTimeList = append(reqTimeList, -1)
			}

		}
		log.Println("ip list:" + strings.Join(ipList, ","))
		log.Println(reqTimeList)
		table.Refresh()

	})
	top := container.New(layout.NewVBoxLayout(), label, inputLayout, editBtn)
	return container.NewBorder(top, nil, nil, nil, table)
}
func GetRequestTime(ip string) int {

	log.Println("get for ip:" + ip)
	conn, err := net.DialTimeout("ip4:icmp", ip, 5*time.Second)
	if err != nil {
		log.Println("yy")

		fmt.Println("无法连接到目标主机:", err)
		return -1
	}
	// 设置读取的超时时间
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.Close()
	var msg [512]byte
	msg[0] = 8  // Echo Request (ping)
	msg[1] = 0  // Code 0
	msg[2] = 0  // Checksum, fix later
	msg[3] = 0  // Checksum, fix later
	msg[4] = 0  // Identifier[0]
	msg[5] = 13 // Identifier[1] (arbitrary)
	msg[6] = 0  // Sequence[0]
	msg[7] = 37 // Sequence[1] (arbitrary)
	len := 8

	check := checkSum(msg[0:len])

	msg[2] = byte(check >> 8)
	msg[3] = byte(check & 255)

	start := time.Now()
	_, err = conn.Write(msg[0:len])
	if err != nil {
		fmt.Println("写入数据失败:", err)
		return -2
	}

	log.Println("xxx11")

	_, err = conn.Read(msg[0:])
	if err != nil {
		fmt.Println("读取数据失败:", err)
		return -3
	}
	log.Println("xxx22")

	duration := time.Since(start).Milliseconds()
	fmt.Printf("Ping成功，延迟时间：%d\n", duration)
	return int(duration)
}

func checkSum(msg []byte) uint16 {
	sum := 0
	// assume even for now
	for n := 0; n < len(msg); n += 2 {
		sum += int(msg[n])*256 + int(msg[n+1])
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var answer uint16 = uint16(^sum)
	return answer
}

// func getRequestTime(ip string) int {
// 	start := time.Now()
// 	address := net.JoinHostPort(ip, "80")

// 	dialer := &net.Dialer{
// 		Timeout: 5 * time.Second,
// 	}
// 	_, err := dialer.Dial("tcp", address)
// 	end := time.Now()
// 	if err == nil {
// 		return -1
// 	} else {
// 		return (int)(end.Sub(start).Milliseconds())
// 	}

// }
