package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
)

var (
	blobDir   string
	Validator = validator.New()
)

type ChunkRequest struct {
	// ChunkIndex specifies the order of the chunk
	// might be used if accepting multiple chunks concurrently
    // and then wirte them to the file with correct order
	ChunkIndex uint      `json:"chunkIndex" validate:"required"`
	FileID     uuid.UUID `json:"fileId" validate:"required,uuid	"`
	Data       []byte    `json:"data" validate:"required"`
	TotaChunks uint      `json:"totalChunks" validate:"required"`
}

func main() {
	setBlobDir()

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

	if err := Validator.Struct(req); err != nil {
		return fiber.ErrBadRequest
	}

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

func setBlobDir() {
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
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("server <BLOB_DIR>")
	fmt.Println("    BLOB_DIR    the dir in which uploaded files will be stored")
}
