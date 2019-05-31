//go:generate go run -tags generate gen.go

package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/zserge/lorca"
)

// Application Resources
type AppRes struct {
	UI lorca.UI
	JitterMS time.Duration
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
	if err := ui.Bind("FetchAllData", func(aid string) string { return FetchAllData(res, aid) }); err != nil { log.Fatal(err) }
	if err := ui.Bind("FetchArticlesData", func(aid []string) string { return FetchArticlesData(res, aid) }); err != nil { log.Fatal(err) }
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
	res.JitterMS = time.Duration(jitterMs) * time.Millisecond
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

type TmpNode struct{
	PMID string
	Id string `json:"id"`
	Title string `json:"title"`
}
type TmpLink struct{
	Source int `json:"source"`
	Target int `json:"target"`
}
type CPack struct {
	Nodes []TmpNode `json:"nodes"`
	Links []TmpLink `json:"links"`
}

func FetchArticlesData(res AppRes, aid []string) string {
	log.Printf("Article IDs : %s\n", aid)

	nodes := make([]TmpNode, 0)

	// links => using node idx.  { "source": 0, "target": 1 }
	var tmpLinks [][]string

	// Fetch all links
	for _, rawid := range aid {
		nodes = append(nodes, TmpNode{
			PMID: rawid,
			Id: fmt.Sprintf("pmid:%s", rawid),
			Title: fmt.Sprintf("No Title Fetch for id %s", rawid),
		})

		if n, err := strconv.Atoi(rawid); err == nil {
			time.Sleep(res.JitterMS)
			for _, ref := range GetPubMedIDsCitedIn(int64(n)) {
				dstID := strconv.FormatInt(ref, 10)
				lk := []string {rawid, dstID}
				tmpLinks = append(tmpLinks, lk)
				fmt.Printf("%d, \"cited in\", %d\n", n, ref)
				nodes = append(nodes, TmpNode{
					PMID: dstID,
					Id: fmt.Sprintf("pmid:%s", dstID),
					Title: fmt.Sprintf("No Title Fetch for id %s", dstID),
				})
				time.Sleep(res.JitterMS)
			}
		}
	}

	id2idx := make(map[string]int)
	// Convert all nodes to map
	for idx, val := range nodes {
		id2idx[val.PMID] = idx
	}

	var tLinks []TmpLink
	// convert tmps links to actual links
	for _, arr := range tmpLinks {
		src := arr[0]
		dst := arr[1]

		tLinks = append(tLinks, TmpLink{
			Source: id2idx[src],
			Target: id2idx[dst],
		})
	}

	cpack := CPack {
		Nodes: nodes,
		Links: tLinks,
	}

	if b, err := json.Marshal(cpack); err == nil {
		return string(b)
	} else {
		return "{\"nodes\": [], \"links\": []}"
	}
}

func GetPubMedIDsCitedIn(pubMedID int64) []int64 {
	resp, err := http.Get(EUtilsGetCitationsUrl(pubMedID)) // Example 4423606
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var results ELinkResult
	if err := xml.Unmarshal(body, &results); err != nil {
		fmt.Printf("Status Code %d\n", resp.StatusCode)
		log.Fatal(err)

	}

	ids := make([]int64, len(results.LinkSet.LinkSetDb.Link))
	for i, id := range results.LinkSet.LinkSetDb.Link {
		ids[i] = id.Id
	}

	return ids
}

func EUtilsGetCitationsUrl(articleID int64) string {
	return fmt.Sprintf("https://eutils.ncbi.nlm.nih.gov/entrez/eutils/elink.fcgi?dbfrom=pubmed&linkname=pubmed_pubmed_citedin&id=%d", articleID)
}

type ELinkResult struct { LinkSet LinkSet }
type Link struct{ Id int64 }
type LinkSetDb struct {
	DbTo     string
	LinkName string
	Link     []Link
}
type LinkSet struct {
	DbFrom    string
	IdList    []Link
	LinkSetDb LinkSetDb
}
