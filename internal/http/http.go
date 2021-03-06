// Copyright 2014 The Cayley Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/amansx/cayley/clog"
	"github.com/gobuffalo/packr/v2"
	"github.com/julienschmidt/httprouter"

	"github.com/amansx/cayley/graph"
	"github.com/amansx/cayley/internal/gephi"
	cayleyhttp "github.com/amansx/cayley/server/http"
)

var static = packr.New("Static", "../../static")

type statusWriter struct {
	http.ResponseWriter
	code *int
}

func (w *statusWriter) WriteHeader(code int) {
	*(w.code) = code
}

func LogRequest(handler httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		start := time.Now()
		addr := req.Header.Get("X-Real-IP")
		if addr == "" {
			addr = req.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = req.RemoteAddr
			}
		}
		code := 200
		rw := &statusWriter{ResponseWriter: w, code: &code}
		clog.Infof("started %s %s for %s", req.Method, req.URL.Path, addr)
		handler(rw, req, params)
		clog.Infof("completed %v %s %s in %v", code, http.StatusText(code), req.URL.Path, time.Since(start))
	}
}

func jsonResponse(w http.ResponseWriter, code int, err interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"error": `))
	data, _ := json.Marshal(fmt.Sprint(err))
	w.Write(data)
	w.Write([]byte(`}`))
}

type API struct {
	config *Config
	handle *graph.Handle
}

func (api *API) GetHandleForRequest(r *http.Request) (*graph.Handle, error) {
	return cayleyhttp.HandleForRequest(api.handle, "single", nil, r)
}

func (api *API) RWOnly(handler httprouter.Handle) httprouter.Handle {
	if api.config.ReadOnly {
		return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
			jsonResponse(w, http.StatusForbidden, "Database is read-only.")
		}
	}
	return handler
}

func CORSFunc(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if origin := req.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
}

func CORS(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		CORSFunc(w, req, params)
		h(w, req, params)
	}
}

func (api *API) APIv1(r *httprouter.Router) {
	r.POST("/api/v1/query/:query_lang", CORS(LogRequest(api.ServeV1Query)))
	r.POST("/api/v1/shape/:query_lang", CORS(LogRequest(api.ServeV1Shape)))
	r.POST("/api/v1/write", CORS(api.RWOnly(LogRequest(api.ServeV1Write))))
	r.POST("/api/v1/write/file/nquad", CORS(api.RWOnly(LogRequest(api.ServeV1WriteNQuad))))
	r.POST("/api/v1/delete", CORS(api.RWOnly(LogRequest(api.ServeV1Delete))))
}

type Config struct {
	ReadOnly bool
	Timeout  time.Duration
	Batch    int
}

func SetupRoutes(handle *graph.Handle, cfg *Config) error {
	r := httprouter.New()
	api := &API{config: cfg, handle: handle}
	r.OPTIONS("/*path", CORSFunc)
	api.APIv1(r)

	api2 := cayleyhttp.NewAPIv2(handle)
	api2.SetReadOnly(cfg.ReadOnly)
	api2.SetBatchSize(cfg.Batch)
	api2.SetQueryTimeout(cfg.Timeout)
	api2.RegisterOn(r, CORS, LogRequest)

	gs := &gephi.GraphStreamHandler{QS: handle.QuadStore}
	const gephiPath = "/gephi/gs"
	r.GET(gephiPath, CORS(gs.ServeHTTP))

	r.GET("/docs/:docpage", serveDocPage)

	setupUI()

	r.GET("/", serveUI)
	r.GET("/ui/:ui_type", serveUI)

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(static)))

	http.Handle("/", r)
	return nil
}
