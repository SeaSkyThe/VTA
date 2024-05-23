package webcam

import (
	"context"
	"log"
	"net/http"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"

	"github.com/gorilla/websocket"
)

func StartStreaming(deviceName string, width uint32, height uint32, endpoint string, port string, frames chan []byte) {
	camera, err := device.Open(deviceName, device.WithPixFormat(
		v4l2.PixFormat{PixelFormat: v4l2.PixelFmtRGB24, Width: width, Height: height},
	))
	if err != nil {
		log.Fatalf("Failed to open device: %s - %s", deviceName, err)
	}

	defer camera.Close()

	if err := camera.Start(context.TODO()); err != nil {
		log.Fatalf("Camera start: %s", err)
	}

	for frame := range camera.GetOutput() {
		frames <- frame
	}

	close(frames)
}

func StreamImage(w http.ResponseWriter, req *http.Request, frames chan []byte) {

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection to Websocket", http.StatusInternalServerError)
		return
	}

	defer conn.Close()

	for frame := range frames {
		if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
			log.Printf("Failed to write message to WebSocket connection: %s", err)
			return
		}
	}

}

func StreamAsciiArt(w http.ResponseWriter, req *http.Request, frames chan string) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection to Websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	for frame := range frames {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(frame)); err != nil {
			return
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
}
