package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dustin/go-humanize"
)

func main() {

	var MaxFormBuffer string
	var MaxFilesize string

	// flags declaration using flag package
	flag.StringVar(&MaxFormBuffer, "b", "maxformbuffer", "Specify the ParseMultipartForm maxMemory buffer size")
	flag.StringVar(&MaxFilesize, "s", "maxfilesize", "Specify the maximumfilesize")

	flag.Parse()

	buffer, err := humanize.ParseBytes(MaxFormBuffer)
	if err != nil {
		fmt.Printf("parsing maxformbuffer: %v\n", err)
		os.Exit(1)
	}

	size, err1 := humanize.ParseBytes(MaxFilesize)
	if err != nil {
		fmt.Printf("parsing maxfilesize: %v\n", err1)
		os.Exit(2)
	}

	fmt.Printf("maxformbuffer: %d\n", int64(buffer))
	fmt.Printf("maxfilesize: %d\n", int64(size))

	// handle route using handler function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		r.Body = http.MaxBytesReader(w, r.Body, int64(size))

		if max_size_err := r.ParseMultipartForm(int64(buffer)); max_size_err != nil {
			fmt.Printf("Upload error: %+v\n", max_size_err)
		}
		file, handler, ff_err := r.FormFile("myFile")
		if ff_err != nil {
			fmt.Printf("FormFile error: %+v\n", ff_err)
		}
		defer file.Close()

		tempFile, tmpf_err := os.OpenFile("./"+handler.Filename, os.O_RDWR|os.O_CREATE, 0755)

		if tmpf_err != nil {
			fmt.Printf("OpenFile error: %+v\n", tmpf_err)
		}
		defer tempFile.Close()

		fileBytes, io_err := io.Copy(tempFile, file)
		if io_err != nil {
			fmt.Printf("ioCopy error: %+v\n", io_err)
		}
		fmt.Printf("Copied fileBytes: %d\n", fileBytes)
		fmt.Printf("Request: %#v\n", r)
		fmt.Fprintf(w, "Welcome to MultiPartForm test!\n")
		os.Exit(0)
	})

	// listen to port
	http.ListenAndServe(":5050", nil)
}
