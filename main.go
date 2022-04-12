package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main(){
	var address string
	uploadEnabled := flag.Bool("u",false,"enable upload handler")
	flag.StringVar(&address, "a", "127.0.0.1:8080", "Specify address. Default is \"127.0.0.1:8080\"")
	flag.Parse()
	s := Server{
		UploadEnabled: *uploadEnabled,
	}
	err := s.BuildRoutes()
	if err != nil{
		log.Fatalf("error during building routes: %v", err)
	}
	for _, f := range s.Routes{
		http.HandleFunc(fmt.Sprintf("/%v", f.Name), f.Create_handler())	
	}
	if s.UploadEnabled{
		http.HandleFunc("/upload", s.UploadHandler)
	}
	http.HandleFunc("/index.html", s.IndexHandler )
	http.HandleFunc("/", s.IndexHandler)
	log.Printf("started on %v", address)
	err = http.ListenAndServe(address, nil)
	if err != nil{
		log.Fatalf("error during serve http: %v", err)
	}
	// path_list := strings.Split(os.Args[0], "/")
	// exec_file := path_list[len(path_list)-1]
	// routes, indexNotExist, err := filteredFileList(exec_file)
	// if err != nil{
	// 	log.Fatalf("error while reading dir: %v", err)
	// }
	// if indexNotExist || (*indexRewrite){
	// 		log.Print("creating index.html")
	// 		createIndex(routes, template, (*enableUpload))
	// 	}
	
	// for _, f := range routes{
	// 	if f.fName == "index.html"{
	// 			continue
	// 		}
	// 	http.HandleFunc(fmt.Sprintf("/%v", f.fName), create_handler(f))
	// 	// log.Printf("%+v", f)
	// 	}
	// indexhandler := create_handler(route{fName: "index.html",fMime: "text/html"})
	// http.HandleFunc("/index.html", indexhandler )
	// http.HandleFunc("/", indexhandler)
	// if (*enableUpload){
	// 	http.HandleFunc("/upload", uploadHandler)
	// }
	// log.Printf("started on %v", address)
	// err = http.ListenAndServe(address, nil)
	// if err != nil{
	// 	log.Fatalf("error during serve http: %v", err)
	// }
}

/*
TODO:
 - add possibility to use index template (if not template - generate index)
 - add option to upload route
*/