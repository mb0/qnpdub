package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetFlags(0)
	flag.Usage = func() { help(flag.CommandLine.Output()) }
	flag.Parse()
	var err error
	args := flag.Args()[1:]
	switch cmd := flag.Arg(0); cmd {
	case "cat":
		err = doCat(args)
	case "clap":
		err = doClap(args)
	case "sync":
		err = doSync(args)
	case "web":
		err = doWeb(args)
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

func help(w io.Writer) {
	fmt.Fprintf(w, `Usage: qnpdub [<flags>] <command> [<args>]

qnpdub provides tools for creating dub video and collages.

Media flags

   -vod=0
   -aod=0
       Offset first video and audio by duration. You can use HH:MM:SS.mm or S+.mm format.

   -dur=0
       Limits the output duration. You can use HH:MM:SS.mm or S+.mm format.

   -fps=0
       Sets the output frame rate, use 30 or 30/1, and 30000/1001 instead of 29.97

   -dim=0
       Sets the output dimensions, use 720:-2 to scale width to 720px preserving input ratio.

   -yes=false
       Override existing output files.


Media commands

   cat <out> <paths>
        Concatenates and combines media files to output by end-clap and prints the offsets as json.
        Uses fps, scale, voff, aoff, dur flags.

   clap <paths>
        Detects a matching end-clap in media files and prints the result as json.
        Uses fps flag.

   sync <out> <paths>
   	Detects a matching end-clap in the last video and audio and concatenates to output.
	The output uses starts with the first audio stream up to the detected clap.
        Uses fps, scale flags.



Other commands

   web
       Starts a local webserver with some information.
       -addr=localhost:8403
           Configures the server address.

   help
       Displays this help message.
`)
}

func doWeb(args []string) error {
	var addr string
	flags := flag.NewFlagSet("web", flag.ContinueOnError)
	flags.StringVar(&addr, "addr", "localhost:8403", "httpd server address")
	flags.Parse(args)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})
	fmt.Printf("Starting mb0/qnpdub server at http://%s", addr)
	return http.ListenAndServe(addr, nil)
}
