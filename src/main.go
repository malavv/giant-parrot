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

type AppRes struct {
	UI lorca.UI
	JitterMS int
}

func main() {
	// Start Program
	var args []string
	ui, err := lorca.New("", "", 1016, 1039, args...)
	if err != nil { log.Fatal(err) }
	defer ui.Close()

	res := AppRes{
		UI: ui,
	}

	// Make Go Functions available

	// Launched on body load by the JS
	if err := ui.Bind("OnAppStarting", func() { OnAppStarting(res) }); err != nil { log.Fatal(err) }
	// Called by JS to get Article Data from Pubmed
	if err := ui.Bind("FetchArticleData", func(aid string) string { return FetchArticleData(res, aid) }); err != nil { log.Fatal(err) }
	if err := ui.Bind("FetchAllData", func(aid string) string { return FetchAllData(res, aid) }); err != nil { log.Fatal(err) }
	if err := ui.Bind("ChangeJitter", func(jitterMs int) { ChangeJitter(res, jitterMs) }); err != nil { log.Fatal(err) }

	// Open FS resource (for CSS and stuff)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil { log.Fatal(err) }
	defer ln.Close()
	go http.Serve(ln, http.FileServer(FS))
	ui.Load(fmt.Sprintf("http://%s", ln.Addr()))

	// Wait until the interrupt signal arrives or browser window is closed
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
		case <-sigc:
		case <-ui.Done():
	}
}

func ChangeJitter(res AppRes, jitterMs int) {
	fmt.Println("Changing Jitter to ", jitterMs)
	res.JitterMS = jitterMs
}

func OnAppStarting(res AppRes) {
	res.UI.Eval(`console.log('Server is ready')`)
	InitApp(res)
}

func InitApp(res AppRes) {
	res.UI.Eval("init()")
}

func FetchAllData(res AppRes, aid string) string {
	log.Printf("Article ID : %s\n", aid)
	content, err := ioutil.ReadFile("data.json")
	if err != nil { log.Fatal(err) }
	return string(content)
}

func FetchArticleData(res AppRes, aid string) string {
	log.Printf("Article ID : %s\n", aid)
	content, err := ioutil.ReadFile("data.json")
	if err != nil { log.Fatal(err) }
	return string(content)
}
