package main

import (
	"context"
	"log"
	"net/http"

	"github.com/altlimit/jsonrpc2"
)

type Calculator struct {
}

func (c *Calculator) Add(ctx context.Context, a int, b int) int {
	return a + b
}

func (c *Calculator) Subtract(ctx context.Context, a float64, b float64) float64 {
	return a - b
}

func (c *Calculator) Divide(a float64, b float64) (float64, error) {
	if b == 0 {
		return 0, jsonrpc2.ServerError{Code: -32001, Data: "divide by zero"}
	}
	return a / b, nil
}

func main() {
	http.Handle("/rpc", jsonrpc2.NewServer(&Calculator{}))
	port := "8090"
	log.Println("Listening", port)
	http.ListenAndServe(":"+port, nil)
}
