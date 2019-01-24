//go:generate go run -tags generate gen.go
package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/jacobsa/go-serial/serial"
	"github.com/zserge/lorca"
	//	"golang.org/x/text/encoding/simplifiedchinese"
)

func dis() string {
	return "hello world!"
}

var defaultCom io.ReadWriteCloser
var hexEnable string

func hexRecieveSet(b string) {
	log.Println("get hex opt", b)
	hexEnable = b
}
func New(opt serial.OpenOptions) bool {
	var err error
	log.Println(opt)
	opt.MinimumReadSize = 4
	defaultCom, err = serial.Open(opt)

	if err != nil {
		log.Println(err)
		return false

	}

	log.Println("serila is opened")
	return true
}

func closeSerial() bool {
	if err := defaultCom.Close(); err != nil {
		return false
	} else {
		return true
	}
}

type sdata struct {
	Type byte
	Data string
}

func send(data sdata) {
	log.Println(data)
	var bdata []byte
	if data.Type == 1 { //16进制
		str := strings.Replace(data.Data, " ", "", -1)
		var err error
		bdata, err = hex.DecodeString(str)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		bdata = []byte(data.Data)
	}

	if _, err := defaultCom.Write(bdata); err != nil {
		log.Println(err)
		return
	}
}

//字符串每隔N个字符用某个字符串分割
func SpliteN(src string, sep string, n int) (des string) {
	if len(src) == 0 {
		return ""
	}
	tempN := len(src) / n
	tempN2 := len(src) % n

	temp := make([]string, tempN+1)
	if tempN == 0 {
		return src
	}
	var temp2 = make([]string, 0)
	for k, _ := range temp {
		if k == tempN {
			temp2 = append(temp2, src[n*k:(n*k+tempN2)])
			temp2 = append(temp2, sep)
			break
		}
		temp2 = append(temp2, src[n*k:(n*k+n)])
		temp2 = append(temp2, sep)
	}
	for _, v := range temp2 {
		des += v
	}
	return des
}

func recieve(ui lorca.UI) {
	for {
		data := make([]byte, 1024)
		if defaultCom == nil {
			continue
		}
		n, err := defaultCom.Read(data)
		if err != nil {
			log.Println(err)
			continue
		}

		if n > 0 {
			log.Println(data[:n])
			//			bs, _ := simplifiedchinese.GBK.NewDecoder().Bytes(data[:n])
			tdata := string(data[:n])
			log.Println(tdata)
			if hexEnable == "true" {
				tdata = hex.EncodeToString(data[:n])
				tdata = SpliteN(tdata, " ", 2)
			}
			//			ui.Eval(`document.getElementById('receiverText').value=isdhkj`)
			log.Println(fmt.Sprintf(`document.getElementById('receiverText').value =%v`, tdata))
			ui.Eval(fmt.Sprintf(`document.getElementById('receiverText').value =
			document.getElementById('receiverText').value+'%s'`, tdata))
		}
	}
}

func main() {
	ui, err := lorca.New("", "", 1000, 700)
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	// A simple way to know when UI is ready (uses body.onload even in JS)
	ui.Bind("start", func() {
		log.Println("UI is ready")
	})
	ui.Bind("testdis", dis)
	ui.Bind("openSerial", New)
	ui.Bind("closeSerial", closeSerial)
	ui.Bind("send", send)
	ui.Bind("hexRecieveSet", hexRecieveSet)
	// Load HTML.
	// You may also use `data:text/html,<base64>` approach to load initial HTML,
	// e.g: ui.Load("data:text/html," + url.PathEscape(html))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go http.Serve(ln, http.FileServer(FS))
	ui.Load(fmt.Sprintf("http://%s", ln.Addr()))
	go recieve(ui)
	// You may use console.log to debug your JS code, it will be printed via
	// log.Println(). Also exceptions are printed in a similar manner.
	// ui.Eval(`
	// 	console.log("Hello, world!");
	// 	console.log('Multiple values:', [1, false, {"x":5}]);
	// `)

	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}

	log.Println("exiting...")
}
