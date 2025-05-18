package main

import (
	"log"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/metalpoch/livelite/internal/srt"
	"github.com/metalpoch/livelite/internal/streaming"
)

var live *streaming.Streaming

func init() {
	godotenv.Load()

	url := os.Getenv("LIVEKIT_URL")
	key := os.Getenv("LIVEKIT_API_KEY")
	secret := os.Getenv("LIVEKIT_API_SECRET")
	storageKey := os.Getenv("STORAGE_KEY")
	storageURL := os.Getenv("STORAGE_URL")
	storageSecret := os.Getenv("STORAGE_SECRET")
	storageBucket := os.Getenv("STORAGE_BUCKET")
	storageBucketURL := os.Getenv("STORAGE_BUCKET_URL")

	for _, str := range [8]string{key, secret, url, storageBucket, storageBucketURL, storageKey, storageSecret, storageURL} {
		if str == "" {
			log.Fatalf("error enviroment varables requried.")
		}
	}

	live = streaming.NewStreaming(streaming.Config{
		Key:              key,
		Secret:           secret,
		Url:              url,
		StorageURL:       storageURL,
		StorageKey:       storageKey,
		StorageSecret:    storageSecret,
		StorageBucket:    storageBucket,
		StorageBucketURL: storageBucketURL,
	})
}

func main() {
	go srt.RunServer()
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

	// Create a thumbnail
	app.Post("/thumbnail/:room/:identity", func(c fiber.Ctx) error {
		room, err := url.QueryUnescape(c.Params("room"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		identity, err := url.QueryUnescape(c.Params("identity"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		res, err := live.StreamingThumbnails(room, identity)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(res)
	})

	log.Fatal(app.Listen(":3000"))
}
