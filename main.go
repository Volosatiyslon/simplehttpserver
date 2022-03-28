package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type route struct{
	fName 	string
	fMime	string
}

const MAXSIZE = 1024*1024*1024 //1G

func createIndex(fileList []route, t string, addUpload bool) (error){
	var template string
	os.Remove("index.html")
	template_file, err:=os.ReadFile(t)
	if err == nil && strings.Index(string(template_file), "##REPLACE##")>0 {
		template = string(template_file)
	}else{
		template = "<html><body><div>this index file is auto generated after start</div><p></p>##REPLACE##</body></html>"
	}
	var content string
	for _,f := range fileList{
		if f.fName == ""{
			continue
		}
		content+=fmt.Sprintf("<p><a href=\"./%v\">%v</a></p>", f.fName, f.fName)
	}
	if addUpload{
		content += "<form id=\"form\" enctype=\"multipart/form-data\" action=\"/upload\" method=\"POST\">"
		content += "<input class=\"input file-input\" type=\"file\" name=\"file\" />"
		// content += "<input class=\"input file-input\" type=\"file\" name=\"file\" multiple />"
		content += "<button class=\"button\" type=\"submit\">Submit</button> </form>"
	}
	file := strings.Replace(template,"##REPLACE##", content, 1) 
	err = os.WriteFile("index.html", []byte(file), 0777)
	if err != nil{
		log.Fatalf("err during creating index: %v", err)
	}
	return nil
}

func filteredFileList(exept string) ([]route, bool, error){
	fileList, err := os.ReadDir(".")
	if err != nil{
		return nil, false, fmt.Errorf("err during reading dir: %v", err)
	}
	needIndex := true
	routeList := make([]route,0, len(fileList))
	for _,f := range fileList{
		if f.IsDir(){
			continue
		}
		if f.Name()[0]=='.'{
			continue
		}
		if f.Name() == exept{
			continue
		}
		if f.Name() == "index.html"{
			needIndex = false
		}
		ext := ""
		s:=strings.Split(f.Name(), ".")
		if ext =  mime.TypeByExtension(fmt.Sprintf(".%v",s[len(s)-1])); ext == ""{
			ext = "application/octet-stream"
		}
		routeList = append(routeList, route{
			fName: f.Name(),
			fMime: ext,
		})
	}
	return routeList,needIndex, nil
}

func create_handler(f route)(handler func(w http.ResponseWriter, r *http.Request)){
	handler = func(w http.ResponseWriter, r *http.Request){
		if r.Method != "GET"{
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(400)
			w.Write([]byte("wrong method, only GET allowed"))
			log.Printf("%v, %v, %v, %v, %v", r.RemoteAddr, r.Method, r.URL.Path, "400", w.Header())
		}
		content, err := os.ReadFile(f.fName)
		if err != nil{
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("error during reading file: %v", err)))
		}
		w.Header().Set("Content-Type", f.fMime)
		w.Write(content)
		log.Printf("%v, %v, %v, %v, %v", r.RemoteAddr, r.Method, r.URL.Path, "200", w.Header())

	}
	return
}

func uploadHandler(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MAXSIZE)
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	dst, err := os.Create(fmt.Sprintf("./%d%s.uploaded", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)

}

func main(){
	var address string
	var template string
	indexRewrite := flag.Bool("r",false,"check to rewrite index file")
	enableUpload := flag.Bool("u",false,"enable upload handler")
	flag.StringVar(&address, "a", "127.0.0.1:8080", "Specify address. Default is \"127.0.0.1:8080\"")
	flag.StringVar(&template, "t", "", "Specify index file template, default is empty (index will be selfgenerated)")
	flag.Parse()	
	path_list := strings.Split(os.Args[0], "/")
	exec_file := path_list[len(path_list)-1]
	routes, indexNotExist, err := filteredFileList(exec_file)
	if err != nil{
		log.Fatalf("error while reading dir: %v", err)
	}
	if indexNotExist || (*indexRewrite){
			log.Print("creating index.html")
			createIndex(routes, template, (*enableUpload))
		}
	
	for _, f := range routes{
		if f.fName == "index.html"{
				continue
			}
		http.HandleFunc(fmt.Sprintf("/%v", f.fName), create_handler(f))
		// log.Printf("%+v", f)
		}
	indexhandler := create_handler(route{fName: "index.html",fMime: "text/html"})
	http.HandleFunc("/index.html", indexhandler )
	http.HandleFunc("/", indexhandler)
	if (*enableUpload){
		http.HandleFunc("/upload", uploadHandler)
	}
	log.Printf("started on %v", address)
	err = http.ListenAndServe(address, nil)
	if err != nil{
		log.Fatalf("error during serve http: %v", err)
	}
}

/*
TODO:
 - add possibility to use index template (if not template - generate index)
 - add option to upload route
*/