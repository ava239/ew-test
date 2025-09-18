package main

import (
	"ew/internal/database"
	"ew/pkg/api"
	"ew/pkg/subscriptions"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sirupsen/logrus"
)

func main() {
	db := database.InitDB()

	repo := subscriptions.NewRepo(db)
	validate := validator.New()

	server := api.NewServer(repo, validate)

	webApp := fiber.New(fiber.Config{
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	webApp.Use(logger.New(logger.Config{
		Format: "${time} ${method} ${path} - ${status} - ${latency}\n",
	}))
	webApp.Use(recover.New())

	api.RegisterHandlers(webApp, api.NewStrictHandler(
		server,
		[]api.StrictMiddlewareFunc{},
	))

	go func() {
		logrus.Info("Listening on :" + os.Getenv("HTTP_BIND"))

		if err := webApp.Listen(":" + os.Getenv("HTTP_BIND")); err != nil {
			logrus.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	logrus.Info("Gracefully shutting down...")

	if err := webApp.ShutdownWithTimeout(5 * time.Second); err != nil {
		logrus.Fatalf("Fiber server shutdown error: %v", err)
	}

	logrus.Info("Running cleanup tasks...")

	db.Close()

	logrus.Info("Fiber was successfully shut down.")
}
