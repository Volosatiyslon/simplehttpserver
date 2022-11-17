package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var tSimple = `
	<html><body>
	<div>this index file is auto generated after start</div><p></p>
	{{range .Routes}}
	<p><a href="./{{ .Name }}">/{{ .Name }}</a></p>
	{{ end }}
	</body></html>
	`
var tWithUpload = `
	<html><body>
	<div>this index file is auto generated after start</div><p></p>
	{{range .Routes}}
	<p><a href="./{{ .Name }}">/{{ .Name }}</a></p>
	{{ end }}
	<form id="form" enctype="multipart/form-data" action="/upload" method="POST">
		<input class="input file-input" type="file" name="file">
		<button class="button" type="submit">Submit</button> 
	</form>
	</body></html>
	`

const MAXSIZE = 1024 * 1024 * 1024 //1G

type Route struct {
	Name  string
	fMime string
}

type Server struct {
	Routes        []Route
	UploadEnabled bool
}

func (f Route) Create_handler() (handler func(w http.ResponseWriter, r *http.Request)) {
	handler = func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(400)
			w.Write([]byte("wrong method, only GET allowed"))
			log.Printf("%v, %v, %v, %v, %v", r.RemoteAddr, r.Method, r.URL.Path, "400", w.Header())
		}
		content, err := os.ReadFile(f.Name)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("error during reading file: %v", err)))
		}
		w.Header().Set("Content-Type", f.fMime)
		w.Write(content)
		log.Printf("%v, %v, %v, %v, %v", r.RemoteAddr, r.Method, r.URL.Path, "200", w.Header())

	}
	return handler
}
func (s *Server) GetRoute(filename string) (Route, error) {
	for _, r := range s.Routes {
		if r.Name == filename {
			return r, nil
		}
	}
	return Route{}, fmt.Errorf("there is no such file")
}

func (s *Server) BuildRoutes() error {
	fileList, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("cant read directory: %v", err)
	}
	s.Routes = make([]Route, 0, len(fileList))
	for _, f := range fileList {
		if f.IsDir() {
			continue
		}
		if f.Name()[0] == '.' {
			continue
		}
		ext := ""
		nameParts := strings.Split(f.Name(), ".")
		if ext = mime.TypeByExtension(fmt.Sprintf(".%v", nameParts[len(nameParts)-1])); ext == "" {
			ext = "application/octet-stream"
		}

		s.Routes = append(s.Routes, Route{
			Name:  f.Name(),
			fMime: ext,
		})
	}
	return nil
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(400)
		w.Write([]byte("wrong method, only GET allowed"))
		log.Printf("%v, %v, %v, %v, %v", r.RemoteAddr, r.Method, r.URL.Path, "400", w.Header())
		return
	}
	content, err := s.ReturnIndex()
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(500)
		w.Write([]byte("internal server error"))
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write(content)
	log.Printf("%v, %v, %v, %v, %v", r.RemoteAddr, r.Method, r.URL.Path, "200", w.Header())
	return
}

func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%+v", r)
	if r.Method != "POST" {
		http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
		log.Printf("Method not allowed: %v", r.Method)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MAXSIZE)
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	newName := fmt.Sprintf("%d%s.uploaded", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
	dst, err := os.Create(fmt.Sprintf("./%v", newName))
	if err != nil {
		http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
		log.Printf("error during reading file: %v", err)
		return
	}
	defer dst.Close()
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
		log.Printf("error during reading file: %v", err)
		return
	}
	log.Printf("newname: %v", newName)
	s.BuildRoutes()
	for _, v := range s.Routes {
		log.Printf("file: %v", v.Name)
	}
	route, err := s.GetRoute(newName)
	log.Printf("%+v", route)
	if err != nil {
		log.Printf("err: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(500)
		w.Write([]byte("internal server error"))
		return
	}
	http.HandleFunc(fmt.Sprintf("/%v", newName), route.Create_handler())
	fmt.Fprintf(w, "File Uploaded\n")
	// http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
}

func (s *Server) ReturnIndex() ([]byte, error) {
	t := template.New("index")
	var err error
	if s.UploadEnabled {
		t, err = t.Parse(tWithUpload)
	} else {
		t, err = t.Parse(tSimple)
	}
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	err = t.Execute(&b, s)
	return b.Bytes(), err
}
