// main_test.go
package main

import (
	"bufio"
	"fmt"
	"io"
	. "net/http"
	"os"
	"testing"
)

func TestMain(t *testing.T) {

	type readCloser struct {
		io.Reader
		io.Closer
	}

	file, err := os.Open("10g.img")

	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	filesize := fileinfo.Size()
	buffer := make([]byte, filesize)

	// bytesread
	_, err2 := file.Read(buffer)
	if err2 != nil {
		fmt.Println(err)
		return
	}

	req := &Request{
		Method: "POST",
		Header: Header{"Content-Type": {`multipart/form-data; boundary="foo123"`}},
		Body:   &readCloser{bufio.NewReader(file), file},
	}
	err1 := req.ParseMultipartForm(1000000000)
	if err1 == nil {
		t.Error("expected multipart EOF, got nil")
	}

	req.Header = Header{"Content-Type": {"text/plain"}}
	err = req.ParseMultipartForm(25)
	if err != ErrNotMultipart {
		t.Error("expected ErrNotMultipart for text/plain")
	}
}
