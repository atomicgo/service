package main

import (
	"net/http"
	"os"

	"atomicgo.dev/service"
)

func main() {
	svc := getService()
	if err := svc.Start(); err != nil {
		svc.Logger.Error("failed to start service", "error", err)
		os.Exit(1)
	}
}

func getService() *service.Service {
	svc := service.New("unit-testing", nil)

	svc.HandleFunc("/hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		_, _ = w.Write([]byte("Hello, " + name + "!"))
	})

	return svc
}
