package main

import (
	"fmt"
	"os"
	"net/http"
	"strings"
	"errors"
	"strconv"
	"time"
	"io"
	goprop "github.com/xing4git/goprop"
)

const (
	listenAddr string = ":9090"
)

var (
	uploadContent string = "<!doctype html>\n" + "<html>\n" + "<head>\n" + "<meta charset='utf-8'>\n" + "<title>upload</title>\n" + "</head>\n" + "<body>\n" + "<form method='POST' action='/upload' enctype='multipart/form-data'>\n" + "choose file to upload: <input name='file' id='file' type='file' /><br />\n" + "<input type='submit' value='upload' />\n" + "</form>\n" + "</body>\n" + "</html>\n"
	basedir       string
	uploaddir     string
)

func main() {
	readConf()

	http.HandleFunc("/local", safeHanlder(localFileHandler))
	http.HandleFunc("/upload", safeHanlder(uploadHandler))
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

func readConf() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: local_file_server [conf path]")
		os.Exit(1)
	}

	confpath := os.Args[1]
	kv, err := goprop.Load(confpath)
	checkErr(err)
	fmt.Println(kv)
	var contains bool = false
	if basedir, contains = kv["basedir"]; !contains {
		checkErr(errors.New("There is no basedir configuration in " + confpath))
	}
	if uploaddir, contains = kv["uploaddir"]; !contains {
		checkErr(errors.New("There is no uploaddir configuration in " + confpath))
	}

	checkDir(basedir)
	checkDir(uploaddir)
	if !strings.HasSuffix(basedir, "/") {
		basedir = basedir + "/"
	}
	if !strings.HasSuffix(uploaddir, "/") {
		uploaddir = uploaddir + "/"
	}

	fmt.Println("visit localhost" + listenAddr + "/upload to upload file.")
	fmt.Println("visit localhost" + listenAddr + "/local?path=[filepath] to visit local file")
}

func checkDir(dirpath string) {
	fileinfo, err := os.Stat(dirpath)
	checkErr(err)
	if !fileinfo.IsDir() {
		checkErr(errors.New(dirpath + " is not a dir"))
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
}

func localFileHandler(w http.ResponseWriter, r *http.Request) {
	checkServerError(r.ParseForm())
	filename := basedir + r.FormValue("path")
	fileinfo, err := os.Stat(filename)
	checkServerError(err)
	if fileinfo.IsDir() {
		checkServerError(errors.New(filename + " is a dir!"))
	}

	// download http header
	w.Header().Set("Content-type", "application/octet-stream")
	w.Header().Set("Content-disposition", "attachment; filename="+fileinfo.Name())
	w.Header().Set("Content-Length", strconv.Itoa(int(fileinfo.Size())))
	http.ServeFile(w, r, filename)
	fmt.Fprintln(os.Stdout, "write file to client: "+filename)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Set("Content-type", "text/html;charset=utf-8")
		io.Copy(w, strings.NewReader(uploadContent))
		fmt.Fprintln(os.Stdout, "write upload content to client")
		return
	} else if r.Method == "POST" {
		f, h, err := r.FormFile("file")
		checkServerError(err)
		defer f.Close()
		destFile := uploaddir + strconv.Itoa(time.Now().Nanosecond()) + "_" + h.Filename
		t, err := os.Create(destFile)
		checkServerError(err)
		defer t.Close()
		_, err = io.Copy(t, f)
		checkServerError(err)
		fmt.Fprintln(w, "upload success...")
		fmt.Fprintln(os.Stdout, "write file to "+destFile)
	}
}

func checkServerError(err error) {
	if err != nil {
		panic(err)
	}
}

/**
 * make an http.HandleFunc be an safe http.HandleFunc
 */
func safeHanlder(httpHanlder http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err, ok := recover().(error); ok {
				fmt.Fprintln(os.Stderr, err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		httpHanlder(w, r)
	}
}
