package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
)

const ChunkSize = 700 * 1024 // 700 KB

type ChunkRequest struct {
	ChunkIndex uint      `json:"chunkIndex"` // index specifies the order of the chunk
	FileID     uuid.UUID `json:"fileId"`
	Data       []byte    `json:"data"`
	TotaChunks uint      `json:"totalChunks"`
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "invalid number of arguments.\n")
		printUsage()
		os.Exit(1)
	}
	path := args[0]

	fd, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file `%s` doesn't exist.\n", path)
		os.Exit(1)
	}
	defer fd.Close()

	finfo, _ := fd.Stat()
	nChunks := finfo.Size() / ChunkSize
	if finfo.Size()%ChunkSize != 0 {
		nChunks++
	}

	fileID := uuid.New()

	for i := 1; i <= int(nChunks); i++ {
		data := make([]byte, ChunkSize)
		n, err := fd.Read(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "faield to read chunk idx: %d of total: %d. error: %v\n", i, nChunks, err)
		}

		body := ChunkRequest{
			ChunkIndex: uint(i),
			TotaChunks: uint(nChunks),
			FileID:     fileID,
			Data:       data[:n],
		}

		encodedBody, err := json.Marshal(&body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "faield to encode chunk idx: %d of total: %d. error: %v\n", i, nChunks, err)
		}

		req, err := http.NewRequest(
			"POST",
			"http://localhost:8080/upload",
			bytes.NewBuffer(encodedBody),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "faield to prepare request for chunk idx: %d of total: %d. error: %v\n", i, nChunks, err)
		}
		req.Header.Set("content-type", "application/json")
		for {
			if _, err := http.DefaultClient.Do(req); err == nil {
				break
			} else {
				fmt.Fprintf(os.Stderr, "faield to send chunk idx: %d of total: %d. error: %v\n", i, nChunks, err)
			}
		}

		fmt.Printf("%d of %d chunks sent\n", i, nChunks)
	}

	fmt.Println("\nyour file id is:", fileID)
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("cli <FILE_PATH>")
	fmt.Println("    FILE_PATH    the file you want to upload")
}
