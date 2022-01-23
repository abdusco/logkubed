package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

type App struct {
	bindAddress string
	broker      *LogBroker
}

func NewApp(port int, broker *LogBroker) *App {
	return &App{
		broker:      broker,
		bindAddress: fmt.Sprintf(":%d", port),
	}
}

func (a *App) Serve() error {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	e.GET("/logs", a.streamLogs)

	go a.broker.PumpMessages()
	return e.Start(a.bindAddress)
}

func (a *App) streamLogs(c echo.Context) error {
	namespace := c.QueryParam("namespace")
	pod := c.QueryParam("pod")
	container := c.QueryParam("container")
	if namespace == "" || pod == "" || container == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing query params"})
	}

	src := &LogSource{namespace, pod, container}

	stream, err := a.broker.Subscribe(src)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot stream logs"})
	}

	websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()
		for msg := range stream.messages {
			_, err := conn.Write([]byte(msg))
			if err != nil {
				return
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
