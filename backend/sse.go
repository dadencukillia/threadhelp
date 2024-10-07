package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	cacheMiddleware "github.com/gofiber/fiber/v3/middleware/cache"
	"github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/valyala/fasthttp"
)

type sseServer struct {
	serverClose chan struct{}
	clients []client
	mutex sync.Mutex
}

type client struct {
	writer *bufio.Writer
	ctx *fasthttp.RequestCtx
	sendMessage chan []byte
}

func NewSSEServer() sseServer {
	return sseServer{
		clients: []client{},
		mutex: sync.Mutex{},
	}
}

func (a *client) SendMessage(b []byte) error {
	n, err := a.writer.Write(b)
	if err != nil || n == 0 {
		return fmt.Errorf("%s %s", err, "or n=0")
	}

	if err := a.writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (a *client) DeleteFromList(server *sseServer) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for i, c := range server.clients {
		if c.writer == a.writer {
			server.clients = append(server.clients[:i], server.clients[i+1:]...)
			break
		}
	}
}

func (a *sseServer) FiberMiddleware() func(c fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		ctx := c.Context()

		ctx.SetContentType("text/event-stream")
		ctx.Response.Header.Set("Cache-Control", "no-cache")
		ctx.Response.Header.Set("Connection", "keep-alive")
		ctx.Response.Header.Set("Transfer-Encoding", "chunked")
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Cache-Control")
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")

		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			chMessage := make(chan []byte)
			clientInstance := client{
				writer: w,
				ctx: ctx,
				sendMessage: chMessage,
			}

			a.mutex.Lock()
			a.clients = append(a.clients, clientInstance)
			a.mutex.Unlock()

			for {
				var msg []byte

				select {
				case msg = <-chMessage:
					if err := clientInstance.SendMessage(msg); err != nil {
						logger.Println(err)
						clientInstance.DeleteFromList(a)
						return
					}

					break
				case <-time.After(20 * time.Second):
					if err := clientInstance.SendMessage([]byte("data: ping\n\n")); err != nil {
						logger.Println(err)
						clientInstance.DeleteFromList(a)
						return
					}

					break
				case <-ctx.Done():
					clientInstance.DeleteFromList(a)
					return
				}
			}
		}))

		return nil
	}
}

func (a *sseServer) FiberMiddlewaresSet() []func(c fiber.Ctx) error {
	alwaysTrueFunction := func(c fiber.Ctx) bool {
		return true
	}

	return []func(c fiber.Ctx) error{
		helmet.New(),
		etag.New(etag.Config{
			Next: alwaysTrueFunction,
		}),
		cacheMiddleware.New(cacheMiddleware.Config{
			Next: alwaysTrueFunction,
		}),
		limiter.New(limiter.Config{
			Next: alwaysTrueFunction,
		}),
		a.FiberMiddleware(),
	}
}

func (a *sseServer) SendBytes(b []byte) error {
	sendData := append([]byte("data: "), append(b, []byte("\n\n")...)...)

	newClients := []client{}
	defer func() {
		a.clients = newClients
	}()

	a.mutex.Lock()
	defer a.mutex.Unlock()

	for _, c := range a.clients {
		c.sendMessage <- sendData

		newClients = append(newClients, c)
	}

	return nil
}

func (a *sseServer) SendJSON(jsonData any) error {
	jsonObject, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	return a.SendBytes(jsonObject)
}
