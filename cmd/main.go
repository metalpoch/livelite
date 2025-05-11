package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/metalpoch/livelite/internal/streaming"
)

var live *streaming.Streaming

func init() {
	url := os.Getenv("LIVEKIT_URL")
	key := os.Getenv("LIVEKIT_API_KEY")
	secret := os.Getenv("LIVEKIT_API_SECRET")

	for _, str := range [3]string{key, secret, url} {
		fmt.Println(str)
		if str == "" {
			log.Fatalf("error enviroment varables requried.")
		}
	}

	live = streaming.NewStreaming(streaming.Config{
		Key:    key,
		Secret: secret,
		Url:    url,
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

	log.Fatal(app.Listen(":3000"))
}
