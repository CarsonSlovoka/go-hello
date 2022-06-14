package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"text/template"
)

type Settings struct {
	GuildID string
	Token   string
	AppID   string
	Cleanup bool
}

//go:embed page/*gohtml page/**/*gohtml
var pageFS embed.FS

//go:embed static/img/favicon.ico
var faviconBytes []byte

func SimpleServer(context *Settings) {
	mux := http.NewServeMux()

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// w.Header().Add("Content-Type", "image/svg+xml")
		w.Header().Add("Content-Type", "image/x-icon")
		_, _ = w.Write(faviconBytes)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rootDir := "./page"

		src := path.Join(rootDir, r.URL.Path)

		if path.Ext(src) == "" { // 不要訪問資料夾，讓其訪問index.html
			src = path.Join(src, "index.gohtml")
		}
		/*
			bytes, err := pageFS.ReadFile(src)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			 _, _ = w.Write(bytes)
		*/

		t := template.Must(template.New(filepath.Base(src)).ParseFS(pageFS, src))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		err := t.Execute(w, context)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
		}
	})

	var server http.Server
	Debug := false
	Port := "0"
	if Debug {
		server = http.Server{Addr: "127.0.0.1:" + Port, Handler: mux} // 本機測試用
	} else {
		server = http.Server{Addr: ":" + Port, Handler: mux} // 正式
	}

	listener, err := net.Listen("tcp", server.Addr)
	fmt.Printf("using port number:%s", listener.Addr().String())
	if err != nil {
		panic(err)
	}
	if err = server.Serve(listener); err != nil {
		panic(err)
	}
}

func main() {
	configString := os.Getenv("SETTINGS")
	if configString == "" {
		panic("Not found env SETTINGS")
	}
	var conf Settings
	if err := json.Unmarshal([]byte(configString), &conf); err != nil {
		panic(err)
	}
	beautifulBytes, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", beautifulBytes)

	go SimpleServer(&conf)

	signalClose := make(chan os.Signal, 1)
	signal.Notify(signalClose, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-signalClose
}
