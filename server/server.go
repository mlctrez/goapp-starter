//go:build !wasm

package server

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	brotli "github.com/anargu/gin-brotli"
	"github.com/gin-gonic/gin"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/nullrocks/identicon"
)

//go:embed web/*
var webDirectory embed.FS

func Run() (shutdownFunc func(ctx context.Context) error, err error) {

	address := os.Getenv("ADDRESS")
	if address == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8000"
		}
		address = "localhost:" + port
	}

	var listener net.Listener
	if listener, err = net.Listen("tcp4", address); err != nil {
		return nil, err
	}

	if isDevelopment() {
		fmt.Printf("running on http://%s\n", address)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	engine.Use(gin.Logger(), gin.Recovery(), brotli.Brotli(brotli.DefaultCompression))

	staticHandler := http.FileServer(http.FS(webDirectory))

	engine.GET("/web/logo-192.png", generateIcon)
	engine.GET("/web/logo-512.png", generateIcon)
	engine.GET("/web/:path", gin.WrapH(staticHandler))

	engine.NoRoute(gin.WrapH(BuildHandler()))
	engine.RedirectTrailingSlash = false

	server := &http.Server{Handler: engine}

	go func() {
		serveErr := server.Serve(listener)
		if serveErr != nil && serveErr != http.ErrServerClosed {
			log.Println(err)
		}
	}()

	return server.Shutdown, nil
}

func generateIcon(c *gin.Context) {
	ig, err := identicon.New("github", 5, 3)
	if err != nil {
		panic(err)
	}

	draw, err := ig.Draw(getRuntimeVersion())
	if err != nil {
		panic(err)
	}

	size := 512

	if strings.Contains(c.Request.RequestURI, "192") {
		size = 192
	}

	headers := c.Writer.Header()
	etag := c.Request.Header.Get("If-None-Match")
	if etag == getRuntimeVersion() {
		c.Writer.WriteHeader(http.StatusNotModified)
		return
	}

	headers.Set("Cache-Control", "no-cache")
	headers.Set("ETag", fmt.Sprintf("%q", getRuntimeVersion()))
	headers.Set("Content-Type", "image/png")
	err = draw.Png(size, c.Writer)
	if err != nil {
		app.Log(err)
	}

}

func BuildHandler() *app.Handler {
	return &app.Handler{
		Author:      "TODO",
		Description: "go-app starter",
		Name:        "go-app starter",
		Scripts:     []string{},
		Icon: app.Icon{
			AppleTouch: "/web/logo-192.png",
			Default:    "/web/logo-192.png",
			Large:      "/web/logo-512.png",
		},
		AutoUpdateInterval: autoUpdateInterval(),
		ShortName:          "starter",
		Version:            getRuntimeVersion(),
		Styles:             []string{},
		Title:              "go-app starter",
	}
}
