package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
)

var blobDir string

// TODO: validation
type ChunkRequest struct {
	ChunkIndex uint      `json:"chunkIndex"` // index specifies the order of the chunk
	FileID     uuid.UUID `json:"fileId"`
	Data       []byte    `json:"data"`
	TotaChunks uint      `json:"totalChunks"`
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "invalid number of arguments.")
		printUsage()
		os.Exit(1)
	}
	blobDir = args[0]

	if err := os.MkdirAll(blobDir, os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create blob dir. error: %v", err)
		os.Exit(1)
	}

	server := fiber.New(fiber.Config{
		BodyLimit: 1 * 1024 * 1024, // 1 MB
	})
	server.Use(logger.New())

	server.Post("/upload", handleUpload)

	server.Listen(":8080")
}

func handleUpload(c *fiber.Ctx) error {
	req := ChunkRequest{}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	// TODO: use a buffer for each fielId for to improve performance
	// use buffer.Write(req.FileId, req.Data, isLastChunk bool)
	// is last chunk will make the buffer close the file
	path := filepath.Join(blobDir, req.FileID.String())
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	defer fd.Close()

	_, err = fd.Write(req.Data)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	log.Printf("recieved chunk | %d/%d | %s", req.ChunkIndex, req.TotaChunks, req.FileID)

	return c.SendStatus(fiber.StatusOK)
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("server <BLOB_DIR>")
	fmt.Println("    BLOB_DIR    the dir in which uploaded files will be stored")
}
