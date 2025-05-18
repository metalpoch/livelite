package srt

import (
	"fmt"
	"log"

	"github.com/haivision/srtgo"
)

func RunServer() {
	options := make(map[string]string)
	options["transtype"] = "live"

	sck := srtgo.NewSrtSocket("localhost", 9999, options)
	defer sck.Close()
	if err := sck.Listen(10); err != nil { // 10 conexiones en backlog
		log.Fatalf("SRT listen error: %v", err)
	}
	log.Println("SRT server listening on localhost:9999")

	for {
		s, _, err := sck.Accept()
		if err != nil {
			log.Println("SRT accept error:", err)
			continue
		}
		go func(conn *srtgo.SrtSocket) {
			defer conn.Close()
			buf := make([]byte, 2048)
			for {
				n, err := conn.Read(buf)
				if err != nil || n == 0 {
					break
				}
				fmt.Printf("Received %d bytes\n", n)
			}
		}(s)
	}
}
