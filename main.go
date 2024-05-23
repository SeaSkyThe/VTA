package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/seaskythe/vta/webcam"
)

func main() {
	fmt.Println("Welcome to VTA!")
	http.Handle("/", http.FileServer(http.Dir("./")))

	port := ":9095"
	deviceName := "/dev/video0"
	var width uint32 = 640
	var height uint32 = 480
	endpoint := "/stream"

	frames := make(chan []byte)
	modifiedFrames := make(chan string)

	go webcam.StartStreaming(deviceName, width, height, endpoint, port, frames)

	// Modify frames
	go func() {
		for frame := range frames {
			// Copy the original frame
			modifiedFrame := make([]byte, len(frame))
			copy(modifiedFrame, frame)

			// Transforming the image
			// Read frame and decode it
			img, _ := webcam.DecodeImage(modifiedFrame)
			// Transform frame
			asciiImg := webcam.TransformToASCII(img)

			// Sending the transformed frame to another channel
			modifiedFrames <- asciiImg
		}
		close(modifiedFrames)
	}()

	log.Printf("Starting webcam streaming via websocket: [%s/stream]", port)
	http.HandleFunc(endpoint, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		webcam.StreamImage(w, req, frames)
	})
	log.Printf("Starting webcam streaming modified via websocket: [%s/ascii]", port)
	http.HandleFunc("/ascii", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Received request")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		webcam.StreamAsciiArt(w, req, modifiedFrames)
	})

	log.Fatal(http.ListenAndServe(port, nil))
}
