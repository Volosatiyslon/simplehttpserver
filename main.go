package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
)

type route struct{
	fName 	string
	fMime	string
}


func createIndex(fileList []route) (error){
	
	content := "<html><body><div>this index file is auto generated after start</div><p></p>"
	for _,f := range fileList{
		if f.fName == ""{
			continue
		}
		content+=fmt.Sprintf("<p><a href=\"./%v\">%v</a></p>", f.fName, f.fName)
	}
	content += "</body></html>"
	os.WriteFile("index.html", []byte(content), 0777)
	return nil
}

func filteredFileList() ([]route, bool, error){
	fileList, err := os.ReadDir(".")
	if err != nil{
		return nil, false, fmt.Errorf("err during reading dir: %v", err)
	}
	needIndex := false
	routeList := make([]route, len(fileList))
	for _,f := range fileList{
		if f.IsDir(){
			continue
		}
		if f.Name()[0]=='.'{
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

func main(){
	var address string
	flag.StringVar(&address, "a", "127.0.0.1:8080", "Specify address. Default is \"127.0.0.1:8080\"")
	flag.Parse()
	
	routes, indexExist, err := filteredFileList()
	if err != nil{
		log.Fatalf("error while reading dir: %v", err)
	}
	if indexExist{
			log.Print("creating index.html")
			createIndex(routes)
		}
	for _, f := range routes{
		if f.fName == "" || f.fName == "index.html"{
				continue
			}
		http.HandleFunc(fmt.Sprintf("/%v", f.fName), create_handler(f))
		// log.Printf("%+v", f)
		}
	indexhandler := create_handler(route{fName: "index.html",fMime: "text/html"})
	http.HandleFunc("/index.html", indexhandler )
	http.HandleFunc("/", indexhandler)
	log.Printf("started on %v", address)
	err = http.ListenAndServe(address, nil)
	if err != nil{
		log.Fatalf("error during serve http: %v", err)
	}
}