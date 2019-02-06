package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

func postFile(msg []byte, url string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile("file", "file")
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	if _, err := fileWriter.Write(msg); err != nil {
		return err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	client := http.DefaultClient

	req, err := http.NewRequest("POST", url, bodyBuf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("pinata_api_key", os.Getenv("PINATA_API_KEY"))
	req.Header.Set("pinata_secret_api_key", os.Getenv("PINATA_SECRET_KEY"))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(body))
	return nil
}

// sample usage
func main() {
	url := "https://api.pinata.cloud/pinning/pinFileToIPFS"
	msg := []byte("hello_world_from_handshake")
	if err := postFile(msg, url); err != nil {
		fmt.Println(err)
	}
}
