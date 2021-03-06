package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/chetan/simpleproxy"
)

var (
	bind           = flag.String("bind", "", "Bind local and remote ports (8000:7000)")
	targetAddr     = flag.String("target", "", "Target URL")
	listenAddr     = flag.String("listen", ":9001", "HTTP Listen Address")
	staticFilePath = flag.String("static", "", "Static files path example: /path/:/staticdirectory")
	targetURL      *url.URL
	cmd            *exec.Cmd
)

func parseOpts() {
	flag.Parse()

	if *targetAddr == "" && *bind == "" {
		log.Fatal("must specify -target OR -bind")
	}

	if *targetAddr != "" {
		targetURL, err := url.Parse(*targetAddr)
		if err != nil {
			log.Fatal(err)
		}

		if targetURL.Scheme != "http" && targetURL.Scheme != "https" {
			log.Println(targetURL.Scheme, targetURL.Scheme == "http")
			log.Fatal("target should have protocol, eg: -target http://localhost:8000 ")
		}
	}

	// use bind shorthand
	if *bind != "" {
		if strings.Contains(*bind, ":") {
			s := strings.Split(*bind, ":")
			localPort, err := strconv.Atoi(s[0])
			if err != nil {
				log.Fatal("failed to parse local port:", err)
				os.Exit(1)
			}
			listenAddr = ptr(fmt.Sprintf(":%d", localPort))

			remotePort, err := strconv.Atoi(s[1])
			if err != nil {
				log.Fatal("failed to parse remote port:", err)
				os.Exit(1)
			}
			targetURL = &url.URL{Scheme: "http", Host: fmt.Sprintf("localhost:%d", remotePort)}

		} else {
			remotePort, err := strconv.Atoi(*bind)
			if err != nil {
				log.Fatal("failed to parse remote port:", err)
				os.Exit(1)
			}
			listenAddr = ptr(fmt.Sprintf(":%d", remotePort))
		}
	}
}

func runCommand() {
	args := flag.Args()
	cmd = exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("[*] running command:", cmd)
	err := cmd.Start()
	if err != nil {
		log.Fatal("error starting command: ", err)
	}
}

func main() {
	parseOpts()

	mux := simpleproxy.NewLoggedMux()
	mux.Handle("/", simpleproxy.CreateProxy(*targetURL, ""))

	if *staticFilePath != "" {
		paths := strings.Split(*staticFilePath, ":")
		fs := http.FileServer(http.Dir(paths[1]))
		mux.Handle(paths[0], http.StripPrefix(paths[0], fs))
	}

	if len(flag.Args()) > 0 {
		runCommand()
	}

	fmt.Printf("[*] starting proxy: http://localhost%s -> %s\n\n", *listenAddr, targetURL)
	log.Fatal(http.ListenAndServe(*listenAddr, mux))
}

func ptr(str string) *string {
	return &str
}
