package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/lightsocks/internal/conf"
	"github.com/xmapst/lightsocks/internal/statistic"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"
)

var (
	serverSecret = ""
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type Traffic struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}

func Server(conf conf.Server) {
	if conf.Port == 0 {
		return
	}
	r := chi.NewRouter()
	serverSecret = conf.Token
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:         300,
	})

	r.Use(c.Handler)
	r.Mount("/debug", middleware.Profiler())

	r.Route("/api", func(r chi.Router) {
		r.Use(authentication)
		r.Get("/", hello)
		r.Get("/traffic", traffic)
		r.Mount("/connections", connectionRouter())
		r.Mount("/dns", dnsRouter())
	})

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", conf.Host, conf.Port))
	if err != nil {
		logrus.Errorf("External controller listen error: %s", err)
		return
	}
	logrus.Infof("RESTful API listening at: %s", l.Addr().String())
	go func() {
		if err = http.Serve(l, r); err != nil {
			logrus.Errorf("RESTful API serve error: %s", err)
		}
	}()
}

func authentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if serverSecret == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Browser websocket not support custom header
		if websocket.IsWebSocketUpgrade(r) && r.URL.Query().Get("token") != "" {
			token := r.URL.Query().Get("token")
			if token != serverSecret {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, ErrUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		header := r.Header.Get("Authorization")
		bearer, token, found := strings.Cut(header, " ")

		hasInvalidHeader := bearer != "Bearer"
		hasInvalidSecret := !found || token != serverSecret
		if hasInvalidHeader || hasInvalidSecret {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func hello(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, render.M{
		"Name":   "LightSocks",
		"GOOS":   runtime.GOOS,
		"GOARCH": runtime.GOARCH,
		"Git": map[string]string{
			"Url":    "https://github.com/xmapst/lightsocks.git",
			"Branch": "master",
		},
	})
}

func traffic(w http.ResponseWriter, r *http.Request) {
	var wsConn *websocket.Conn
	if websocket.IsWebSocketUpgrade(r) {
		var err error
		wsConn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
	}

	if wsConn == nil {
		w.Header().Set("Content-Type", "application/json")
		render.Status(r, http.StatusOK)
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	t := statistic.DefaultManager
	buf := &bytes.Buffer{}
	var err error
	for range tick.C {
		buf.Reset()
		up, down := t.Now()
		if err := json.NewEncoder(buf).Encode(Traffic{
			Up:   up,
			Down: down,
		}); err != nil {
			break
		}

		if wsConn == nil {
			_, err = w.Write(buf.Bytes())
			w.(http.Flusher).Flush()
		} else {
			err = wsConn.WriteMessage(websocket.TextMessage, buf.Bytes())
		}

		if err != nil {
			break
		}
	}
}
