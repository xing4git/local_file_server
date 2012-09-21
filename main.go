package main

import (
	"fmt"
	"os"
	"strings"
	"errors"
	"net/http"
	goprop "github.com/xing4git/goprop"
	"html/template"
)

var (
	templateCache map[string]*template.Template
	port          string
	basedir       string
	uploaddir     string
	imgDirpath    string = "/images/dir.png"
	imgFilepath   string = "/images/file.png"
)

func init() {
	templateCache = make(map[string]*template.Template)

	for _, subpath := range []string{"upload", "dir"} {
		t := template.Must(template.ParseFiles("html/" + subpath + ".html"))
		templateCache[subpath] = t
	}
}

func main() {
	readConf()

	http.HandleFunc("/local/", safeHanlder(localFileHandler))
	http.HandleFunc("/upload/", safeHanlder(uploadHandler))
	http.HandleFunc("/images/", safeHanlder(imagesHandler))
	http.HandleFunc("/download/", safeHanlder(downloadHandler))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

/**
 * read basedir, uploaddir, port from conf file.
 */
func readConf() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: local_file_server [conf path]")
		os.Exit(1)
	}

	confpath := os.Args[1]
	kv, err := goprop.Load(confpath)
	checkErr(err)
	fmt.Println(kv)

	// check contains
	var contains bool = false
	if basedir, contains = kv["basedir"]; !contains {
		checkErr(errors.New("There is no basedir configuration in " + confpath))
	}
	if uploaddir, contains = kv["uploaddir"]; !contains {
		checkErr(errors.New("There is no uploaddir configuration in " + confpath))
	}
	if port, contains = kv["port"]; !contains {
		checkErr(errors.New("There is no port configuration in " + confpath))
	}

	checkDir(basedir)
	checkDir(uploaddir)

	// if the last char is '/', delete
	if strings.HasSuffix(basedir, "/") {
		basedir = basedir[:len(basedir)-1]
	}
	if strings.HasSuffix(uploaddir, "/") {
		uploaddir = uploaddir[:len(uploaddir)-1]
	}

	fmt.Printf("basedir = %s, uploaddir = %s, port = %s\n", basedir, uploaddir, port)
}

/**
 * check basedir or uploaddir whether is a dir. 
 * if uploaddir is not exist, create it.
 */
func checkDir(dirpath string) {
	fileinfo, err := os.Stat(dirpath)
	if dirpath == uploaddir && os.IsNotExist(err) {
		err = os.MkdirAll(uploaddir, 0774)
		if err == nil {
			fileinfo, err = os.Stat(dirpath)
		}
	}

	checkErr(err)
	if !fileinfo.IsDir() {
		checkErr(errors.New(dirpath + " is not a dir"))
	}
}

/**
 * if error happens before http server starting, then exit
 */
func checkErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
}