package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/metalpoch/livelite/internal/ffmpeg"
	"github.com/metalpoch/livelite/internal/storage"
	"github.com/metalpoch/livelite/internal/streaming"
	"github.com/metalpoch/livelite/internal/watcher"
)

var storageBucketURL string
var live streaming.Streaming
var bucket storage.Storage

func init() {
	godotenv.Load()
	livekitURL := os.Getenv("LIVEKIT_URL")
	livekitKey := os.Getenv("LIVEKIT_API_KEY")
	livekitSecret := os.Getenv("LIVEKIT_API_SECRET")

	storageURL := os.Getenv("STORAGE_URL")
	storageRegion := os.Getenv("STORAGE_REGION")
	storageBucketKey := os.Getenv("STORAGE_BUCKET_KEY")
	storageBucketName := os.Getenv("STORAGE_BUCKET_NAME")
	storageBucketSecret := os.Getenv("STORAGE_BUCKET_SECRET")
	storageBucketURL = os.Getenv("STORAGE_BUCKET_URL")

	for _, str := range [9]string{livekitKey, livekitSecret, livekitURL, storageURL, storageRegion, storageBucketURL, storageBucketName, storageBucketKey, storageBucketSecret} {
		if str == "" {
			log.Fatalf("error enviroment varables requried.")
		}
	}

	live = streaming.NewStreaming(streaming.Config{
		Key:    livekitKey,
		Secret: livekitSecret,
		Url:    livekitURL,
	})

	bucket = storage.NewStorage(storage.Config{
		Url:          storageURL,
		Region:       storageRegion,
		Bucket:       storageBucketName,
		BucketKey:    storageBucketKey,
		BucketSecret: storageBucketSecret,
	})
}

func main() {
	app := fiber.New()
	app.Use(logger.New())

	// List rooms
	app.Get("/room", func(c fiber.Ctx) error {
		res, err := live.ListRooms()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(res)
	})

	// Create room
	app.Post("/room/:name", func(c fiber.Ctx) error {
		name, err := url.QueryUnescape(c.Params("name"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		res, err := live.CreateRoom(name)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(res)
	})

	// Delete Room
	app.Delete("/room/:room", func(c fiber.Ctx) error {
		room, err := url.QueryUnescape(c.Params("room"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		live.DeleteRoom(room)
		return c.JSON(fiber.StatusOK)
	})

	// Join Room
	app.Get("/join/:room/:identity", func(c fiber.Ctx) error {
		room, err := url.QueryUnescape(c.Params("room"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		identity, err := url.QueryUnescape(c.Params("identity"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		res, err := live.JoinToken(room, identity)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(res)
	})

	// Kick user from room
	app.Delete("/kick/:room/:identity", func(c fiber.Ctx) error {
		room, err := url.QueryUnescape(c.Params("room"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		identity, err := url.QueryUnescape(c.Params("identity"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		live.KickUser(room, identity)
		return c.JSON(fiber.StatusOK)
	})

	// RTMP
	app.Get("/rtmp/:room/:name/:identity", func(c fiber.Ctx) error {
		room, err := url.QueryUnescape(c.Params("room"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		name, err := url.QueryUnescape(c.Params("name"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		identity, err := url.QueryUnescape(c.Params("identity"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		res, err := live.GetRTMP(room, name, identity)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(res)
	})

	// FFmpeg HLS
	app.Post("/hls/:stream", func(c fiber.Ctx) error {
		streamName, err := url.QueryUnescape(c.Params("stream"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		rtmpURL := "rtmp://localhost/live/" + streamName
		hlsDir := "/tmp/hls/" + streamName
		hlsPath := filepath.Join(hlsDir, streamName+".m3u8")
		hlsTime := "10"
		hlsListSize := "18"
		hlsSegmentFilename := filepath.Join(hlsDir, streamName+"_%03d.ts")

		if err := os.MkdirAll(hlsDir, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		filepath := fmt.Sprintf("streaming/vod/%s", streamName)
		go watcher.WatchAndUploadHLS(hlsDir, filepath, bucket, 2*time.Minute)

		if err := ffmpeg.StreamToHLS(rtmpURL, hlsSegmentFilename, hlsPath, hlsTime, hlsListSize); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "HLS VOD started",
			"hls_url": storageBucketURL + "/" + filepath + "/" + streamName + ".m3u8",
		})
	})

	app.Listen(":3000")
}
