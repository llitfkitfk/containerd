package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
)

var (
	flContainerdSocket string
	flListenAddr       string
)

func init() {
	flag.StringVar(&flContainerdSocket, "h", "./var/run/docker/containerd/docker-containerd-debug.sock", "path to the Docker socket")
	flag.StringVar(&flListenAddr, "l", ":8080", "listen address")
}

func main() {
	flag.Parse()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := flContainerdSocket

		conn, err := net.Dial("unix", target)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c := httputil.NewClientConn(conn, nil)
		defer c.Close()

		res, err := c.Do(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer res.Body.Close()

		for k, vv := range res.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		if _, err := io.Copy(w, res.Body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Fatal(http.ListenAndServe(flListenAddr, handler))

}
