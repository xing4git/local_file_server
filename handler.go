package main

import (
	"net/http"
	"strconv"
	"time"
	"io"
	"net/url"
	"fmt"
	"os"
	"strings"
	"path"
)

// used in html template file: dir.html 
type ItemFile struct {
	Imgsrc   string
	Filepath string
	Filename string
	Download bool // file can be downloaded, but dir cannot
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	subpath, fileinfo, err := getFileInfoFromUrl(r.URL, "/download")
	checkServerError(err)
	fullpath := basedir + subpath

	if fileinfo.IsDir() {
		fmt.Fprintln(w, "cannot download dir")
	}

	// download http header
	w.Header().Set("Content-type", "application/octet-stream")
	w.Header().Set("Content-disposition", "attachment; filename="+fileinfo.Name())
	w.Header().Set("Content-Length", strconv.Itoa(int(fileinfo.Size())))
	http.ServeFile(w, r, fullpath)
	fmt.Fprintln(os.Stdout, "client download: "+fullpath)
}

func localFileHandler(w http.ResponseWriter, r *http.Request) {
	subpath, fileinfo, err := getFileInfoFromUrl(r.URL, "/local")
	checkServerError(err)
	fullpath := basedir + subpath

	if fileinfo.IsDir() {
		dir, err := os.Open(fullpath)
		checkServerError(err)
		fis, err := dir.Readdir(0)
		checkServerError(err)

		files := make([]ItemFile, 0, len(fis))
		parent := path.Dir(subpath)
		files = append(files, ItemFile{Imgsrc: imgDirpath, Filepath: parent, Filename: "..", Download: false})
		for _, fi := range fis {
			// exclude hidden file or dir
			if strings.HasPrefix(fi.Name(), ".") {
				continue
			}
			var urlpath string
			if len(subpath) == 1 {
				urlpath = fi.Name()
			} else if strings.HasSuffix(subpath, "/") {
				urlpath = subpath + fi.Name()
			} else {
				urlpath = subpath + "/" + fi.Name()
			}

			var imgsrc string = imgFilepath
			var download bool = true
			if fi.IsDir() {
				imgsrc = imgDirpath
				download = false
			}
			item := ItemFile{Imgsrc: imgsrc, Filepath: urlpath, Filename: fi.Name(), Download: download}
			files = append(files, item)
		}
		locals := make(map[string]interface{})
		locals["files"] = files
		locals["title"] = fullpath
		renderHtml(w, "dir", locals)
	} else {
		http.ServeFile(w, r, fullpath)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)
	if r.Method == "GET" {
		w.Header().Set("Content-type", "text/html;charset=utf-8")
		renderHtml(w, "upload", nil)
		fmt.Fprintln(os.Stdout, "write upload content to client")
		return
	} else if r.Method == "POST" {
		f, h, err := r.FormFile("file")
		checkServerError(err)
		defer f.Close()
		destFile := uploaddir + "/" + strconv.Itoa(time.Now().Nanosecond()) + "_" + h.Filename
		t, err := os.Create(destFile)
		checkServerError(err)
		defer t.Close()
		_, err = io.Copy(t, f)
		checkServerError(err)
		fmt.Fprintln(w, "upload success...")
		fmt.Fprintln(os.Stdout, "write file to "+destFile)
	}
}

func imagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == imgDirpath {
		http.ServeFile(w, r, imgDirpath[1:])
	} else if r.URL.Path == imgFilepath {
		http.ServeFile(w, r, imgFilepath[1:])
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

func getFileInfoFromUrl(rurl *url.URL, prefix string) (string, os.FileInfo, error) {
	fmt.Println("request path = " + rurl.Path)
	subpath := rurl.Path[len(prefix):]
	if subpath == "" {
		subpath = "/"
	}
	subpath, err := url.QueryUnescape(subpath)
	if err != nil {
		return "", nil, err
	}

	fullpath := basedir + subpath
	fileinfo, err := os.Stat(fullpath)
	if err != nil {
		return "", nil, err
	}
	return subpath, fileinfo, nil
}

func renderHtml(w http.ResponseWriter, htmlpath string, locals map[string]interface{}) error {
	return templateCache[htmlpath].Execute(w, locals)
}

/**
 * if error happens during http server, throw panic, and panic will be recovered in the safeHanlder method
 */
func checkServerError(err error) {
	if err != nil {
		panic(err)
	}
}
