package main

import (
	"archive/zip"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"math/big"
	"math/rand"
)

var host = flag.String("h", "127.0.0.1", "host to listen to")
var shareCntTtl = flag.Int("s", 2, "will be shared n times")
var port = flag.String("p", "8005", "port to listen to")

var path string
var filename string
var key string
var count atomic.Uint64

var handler http.Handler

func genRandKey() string {
	entropy := make([]byte, 32)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	_, err := rng.Read(entropy)
	if err != nil {
		panic(err)
	}
	var i big.Int
	i.SetBytes(entropy)
	return fmt.Sprintf("%0"+fmt.Sprint(28)+"s", i.Text(62)) // base62, url safe
}

func zipFolder(w http.ResponseWriter) {
	w.Header().Add("Content-Disposition", "attachment; filename=\""+filename+".zip\"")
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	err := filepath.Walk(path, func(lpath string, f fs.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if f.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(path, lpath)
		if err != nil {
			panic(err)
		}
		header, err := zip.FileInfoHeader(f)
		if err != nil {
			panic(err)
		}
		header.Name = filepath.ToSlash(rel) // make the paths consistent between OSes
		header.Method = zip.Store
		headerWriter, err := zipWriter.CreateHeader(header)
		if err != nil {
			panic(err)
		}
		file, err := os.Open(lpath)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		_, err = io.Copy(headerWriter, file)
		if err != nil {
			panic(err)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}
}

func serve(w http.ResponseWriter, r *http.Request) {
	// observe key at last part of url - practical path based for proxies
	split := strings.Split(r.URL.Path, "/")
	if len(split) < 1 || split[len(split)-1:][0] != key {
		fmt.Printf("%s invalid path\n", r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized) // 401
		return
	}

	now := count.Add(1)
	if now > uint64(*shareCntTtl) {
		// currently being downloaded, waiting to die
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Printf("File accessed, count #%d\n", now)

	f, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	if f.IsDir() {
		zipFolder(w)
	} else {
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		http.ServeFile(w, r, path)
	}

	fmt.Printf("File fully served #%d\n", now)
	if now >= uint64(*shareCntTtl) {
		fmt.Printf("Stopping server\n")
		os.Exit(0)
	}
}

func usage(msg string) {
	if msg != "" {
		fmt.Printf("%s\n", msg)
	}
	fmt.Printf("Tempshare - CLI tool to share a file or folder a limited number of times\n\n")
	fmt.Printf("An unique URL will be generated, and the file/folder will be served at this path\n")
	fmt.Printf("The file/folder will be served a limited number of times, then the server will stop\n")
	fmt.Printf("Files are served directly, folder are zipped on the fly before being served\n")
	fmt.Printf("\nUsage: tempshare [options] <file/folder>\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var err error
	server := &http.Server{Addr: *host + ":" + *port, Handler: handler}
	flag.Parse()
	path = flag.Arg(0)

	if path == "" {
		usage("")
	} else {
		_, err := os.Stat(path)
		if err != nil {
			panic(err)
		}
	}

	filename = filepath.Base(path)

	if *shareCntTtl < 0 {
		panic("requires a share count > 0")
	}

	count.Store(0)
	key = genRandKey()
	http.HandleFunc("/", serve)

	fmt.Printf("Tempshare starting on path %s\n", path)
	fmt.Printf("Listening on http://%s:%s/%s\n", *host, *port, key)
	if err = server.ListenAndServe(); err != http.ErrServerClosed {
		panic(err)
	}
}
