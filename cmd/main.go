package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/metalpoch/livelite/internal/streaming"
)

var storageBucketURL string
var live streaming.Streaming

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
		Key:            livekitKey,
		Secret:         livekitSecret,
		Url:            livekitURL,
		BucketKey:      storageBucketKey,
		BucketSecret:   storageBucketSecret,
		BucketRegion:   storageRegion,
		BucketEndpoint: storageURL,
		BucketName:     storageBucketName,
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

	app.Post("/egress/hls/:room", func(c fiber.Ctx) error {
		room, err := url.QueryUnescape(c.Params("room"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		filenamePrefix := fmt.Sprintf("streaming/%s/%s", room, time.Now().Format("20060102_150405"))
		playlistName := fmt.Sprintf("%s.m3u8", room)
		livePlaylistName := fmt.Sprintf("%s-live.m3u8", room)

		egressRes, err := live.StartRoomHLSegress(room, filenamePrefix, playlistName, livePlaylistName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"egress_id":     egressRes.EgressId,
			"playlist_url":  fmt.Sprintf("%s/%s/%s", storageBucketURL, filenamePrefix, playlistName),
			"live_playlist": fmt.Sprintf("%s/%s/%s", storageBucketURL, filenamePrefix, livePlaylistName),
		})
	})

	app.Delete("/ingress/:id", func(c fiber.Ctx) error {
		id, err := url.QueryUnescape(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		err = live.DeleteIngress(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"result": "ingress deleted"})
	})

	app.Delete("/egress/:id", func(c fiber.Ctx) error {
		id, err := url.QueryUnescape(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		err = live.DeleteEgress(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"result": "egress deleted"})
	})

	app.Listen(":3000")
}
