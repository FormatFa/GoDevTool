package encode

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"

	"fmt"
	"image"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const maxStringShowByte int = 1000

func Base64View(w fyne.Window) fyne.CanvasObject {
	var resultBuff []byte
	input1 := widget.NewMultiLineEntry()
	input1.SetPlaceHolder("@开头从特定路径读取")
	input1.SetMinRowsVisible(5)
	selectBtn := widget.NewButton("或者选择文件", func() {

		dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			input1.SetText("@" + uc.URI().Path())
		}, w).Show()
	})

	input2 := widget.NewMultiLineEntry()
	input2.SetPlaceHolder("@开头输出到特定路径")
	input2.SetMinRowsVisible(5)

	preViewImg := canvas.NewImageFromImage(image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{100, 100}}))
	preViewImg.SetMinSize(fyne.Size{Width: 100, Height: 100})
	var preViewContent *fyne.Container
	previewRadio := widget.NewRadioGroup([]string{"文本", "图片"}, func(s string) {

		if s == "文本" {
			preViewContent.Objects[3] = input2
		} else {

			preViewContent.Objects[3] = preViewImg
		}
		preViewContent.Refresh()
	})
	saveAsBtn := widget.NewButton("另存为", func() {
		dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
			if uc == nil || uc.URI() == nil {
				// cancel
				return
			}
			log.Println("另存到", uc.URI(), " size=", len(resultBuff))
			uc.Write(resultBuff)
			uc.Close()
			dialog.ShowInformation("提示", "保存成功", w)
		}, w)
	})

	combo := widget.NewSelect([]string{"base64", "md5", "json"}, func(value string) {

	})
	// TODO 记住上次选择
	combo.SetSelected("base64")
	preViewContent = container.New(layout.NewVBoxLayout(), widget.NewLabel("预览方式"), previewRadio, saveAsBtn, input2)

	encode := widget.NewButton("加密", func() {
		var err error
		resultBuff, err = process(input1.Text, input2.Text, "encode", combo.Selected)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		log.Println("encode result=", string(resultBuff))
		if combo.Selected == "md5" {
			input2.SetText(fmt.Sprintf("%x", resultBuff))
		} else {
			if len(resultBuff) > maxStringShowByte {
				input2.SetText(string(resultBuff[:maxStringShowByte]) + " 过长已省略,请另存为文件后查看")
			} else {
				input2.SetText(string(resultBuff))
			}
		}

	})
	decode := widget.NewButton("解密", func() {
		resultBuff, err := process(input1.Text, input2.Text, "decode", combo.Selected)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if previewRadio.Selected == "图片" {
			img, format, err := image.Decode(bytes.NewReader(resultBuff))
			log.Println("preview img format=", format)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			preViewImg.Image = img
			preViewImg.Refresh()
		} else {
			if len(resultBuff) > maxStringShowByte {
				input2.SetText(string(resultBuff[:maxStringShowByte]) + " 过长已省略,请另存为文件后查看")
			} else {
				input2.SetText(string(resultBuff))
			}

		}

	})

	btnLayout := container.New(layout.NewHBoxLayout(), encode, decode, combo)

	content := container.New(layout.NewVBoxLayout(), input1, selectBtn, btnLayout, preViewContent)

	return content
}

func process(input, output string, action string, algorithm string) ([]byte, error) {
	log.Println("process input=", input, "output=", output, "action=", action+" algo:"+algorithm)
	var inputBytes []byte
	var resultBuff []byte
	if strings.HasPrefix(input, "@") {
		log.Println("read file => ", input[1:])
		inputBytes, _ = os.ReadFile(input[1:])
	} else {
		inputBytes = []byte(input)
	}
	if action == "encode" {

		if algorithm == "base64" {
			resultBuff = make([]byte, base64.StdEncoding.EncodedLen(len(inputBytes)))
			base64.StdEncoding.Encode(resultBuff, inputBytes)
		} else if algorithm == "md5" {
			resultBuff = make([]byte, 16)
			temp := md5.Sum(inputBytes)
			copy(resultBuff, temp[:])
		} else if algorithm == "json" {
			resultBuff = make([]byte, 16)
			var buffer bytes.Buffer
			err := json.Compact(&buffer, inputBytes)
			if err != nil {
				log.Fatalln("error", err.Error())
			}
			resultBuff = buffer.Bytes()

		}

	}
	if action == "decode" {
		if algorithm == "base64" {
			resultBuff = make([]byte, base64.StdEncoding.DecodedLen(len(inputBytes)))
			base64.StdEncoding.Decode(resultBuff, inputBytes)
		} else if algorithm == "md5" {
			return nil, errors.New("不支持")
		} else if algorithm == "json" {
			log.Println("decode json")
			resultBuff = make([]byte, 16)
			var buffer bytes.Buffer
			err := json.Indent(&buffer, inputBytes, "", " ")
			if err != nil {
				log.Fatalln("error", err.Error())
			}
			resultBuff = buffer.Bytes()
			log.Println("result" + string(resultBuff))
		}

	}

	return resultBuff, nil

}
