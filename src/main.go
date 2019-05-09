//go:generate go run -tags generate gen.go

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/zserge/lorca"
)

func OnData() string {
	content, err := ioutil.ReadFile("data.json")
	if err != nil { log.Fatal(err) }
	return string(content)
}

func main() {
	args := []string{}
	
	ui, err := lorca.New("", "", 1016, 1039, args...)
	if err != nil { log.Fatal(err) }
	defer ui.Close()

	// A simple way to know when UI is ready (uses body.onload event in JS)
	//ui.Bind("start", func() { log.Println("UI is ready") }) // This is supposed to work
	ui.Bind("start", func() { 
		ui.Eval(`console.log('UI is ready')`)
		ui.Eval(`init({width:1000, height: 1000})`)
	})
	ui.Bind("hello", func() { ui.Eval(`console.log('Hello world!')`) })

	ui.Bind("loadData", OnData)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil { log.Fatal(err) }
	defer ln.Close()
	go http.Serve(ln, http.FileServer(FS))
	ui.Load(fmt.Sprintf("http://%s", ln.Addr()))

	// You may use console.log to debug your JS code, it will be printed via
	// log.Println(). Also exceptions are printed in a similar manner.
	ui.Eval(`
		console.log('Multiple values:', [1, false, {"x":5}]);
	`)

	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
		case <-sigc:
		case <-ui.Done():
	}

	log.Println("exiting...")
}


