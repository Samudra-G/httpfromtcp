package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Samudra-G/http-golang/internal/request"
	"github.com/Samudra-G/http-golang/internal/response"
	"github.com/Samudra-G/http-golang/internal/server"
)

const port = 42069

func main() {
	s, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message: "Your problem is not my problem\n",
			}
		case "/myproblem":
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message: "Woopsie, my bad\n",
			}
		default:
			w.Write([]byte("All good, frfr\n"))
		}
		return nil
	})
	
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}