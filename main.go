package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mb0/qnpdub/clap"
)

var addr = flag.String("addr", "localhost:8403", "httpd server address")
var rate = flag.Int("rate", 30, "frame rate to sync to")

func main() {
	flag.Parse()
	var err error
	args := flag.Args()
	switch cmd := flag.Arg(0); cmd {
	case "web":
		err = web(*addr)
	case "sync":
		err = sync(*rate, args[1:])
	case "help":
		help(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "invalid command %s", cmd)
		help(os.Stderr)
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func help(f *os.File) {
	fmt.Fprintf(f, `Usage: qnpdub [<flags>] <command> [<args>]

qnpdub provides tools for creating dub video and collages.

Commands

   sync		Syncs media files by end-clap and prints the offsets as json.
   		Flag <-rate=30> sets the frame-rate to sync to.

   web		Starts a local webserver with some information.
   		Flag <-addr=localhost:8403> configures the server address.

   help		Displays this help message.
`)
}

func sync(rate int, paths []string) error {
	d := clap.Default()
	offs, err := d.SyncPaths(30, paths...)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(offs)
}

func web(addr string) error {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})
	fmt.Printf("Starting mb0/qnpdub server at http://%s", addr)
	return http.ListenAndServe(addr, nil)
}
