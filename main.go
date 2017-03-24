package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

var staticDir string

func init() {
	cwd, _ := os.Getwd()
	staticDir = path.Join(cwd, "static")
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	logrus.Info("GET /")

	index := path.Join(staticDir, "index.html")
	fd, err := os.Open(index)
	if err != nil {
		logrus.Warn(err)
		http.NotFound(w, r)
		return
	}
	defer fd.Close()

	body, err := ioutil.ReadAll(fd)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(body)
	if err != nil {
		logrus.Error(err)
	}
}

func asyncAccept(listener net.Listener) (<-chan net.Conn, error) {
	cc := make(chan net.Conn)

	go func() {
		defer close(cc)
		conn, err := listener.Accept()
		if err == nil {
			logrus.Info("Got connection!")
			cc <- conn
		}
	}()

	return cc, nil
}

func dump(dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	io.Copy(dst, src)
}

func ProxyWire(ws *websocket.Conn, conn net.Conn) {
	logrus.Info("Wiring ", ws.RemoteAddr(), " <=> ", conn.RemoteAddr())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go dump(ws, conn, wg)
	go dump(conn, ws, wg)
	wg.Wait()
	logrus.Info("Wiring terminated")
}

func pingWebSocket(ws *websocket.Conn) <-chan error {
	closed := make(chan error)
	go func() {
		for {
			if _, err := ws.Write([]byte("PING")); err != nil {
				closed <- err
				close(closed)
				return
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return closed
}

func ProxyListen(ws *websocket.Conn) {
	defer ws.Close()
	listenAddress := net.TCPAddr{}

	listener, err := net.ListenTCP("tcp", &listenAddress)
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info("Listening on ", listener.Addr())
	defer listener.Close()

	closed := pingWebSocket(ws)

	for {
		ws.Write([]byte(fmt.Sprint("LISTEN ", listener.Addr())))

		cc, err := asyncAccept(listener)
		if err != nil {
			logrus.Error(err)
			break
		}

		var conn net.Conn
		var ok bool

		select {
		case conn, ok = <-cc:
			if !ok {
				logrus.Warn("Listener failed")
				return
			}
		case closed := <-closed:
			logrus.Warn(closed)
			return
		}

		ProxyWire(ws, conn)
		conn.Close()
	}

	logrus.Info("Proxy finished")
}

func WebSocketProxy(ws *websocket.Conn) {
	logrus.Info("Websocket proxy initiated")
	ProxyListen(ws)
	logrus.Info("Done here")
}

func main() {
	logrus.Info("Serving static assets from ", staticDir)

	http.HandleFunc("/", HomeHandler)
	http.Handle("/socket", websocket.Handler(WebSocketProxy))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir+"/"))))
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}
