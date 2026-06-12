package main

import (
	"embed"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var buildAt string
var gitHash string

func main() {
	logrus.SetReportCaller(true)
	r := gin.Default()
	r.GET("/version", func(c *gin.Context) { c.JSON(200, gin.H{gitHash: buildAt}) })
	r.Use(feMw("/"))
	r.GET("/ws", ApiWsGuacamole())
	// HTTPS
	logrus.Info("Starting HTTPS server on :9528")
	logrus.Fatal(r.RunTLS(":9528", "certs/cert.pem", "certs/key.pem"))
}

//go:embed frontend/dist/*
var fs embed.FS

const fsBase = "frontend/dist"

func feMw(urlPrefix string) gin.HandlerFunc {
	const indexHtml = "index.html"

	return func(c *gin.Context) {
		urlPath := strings.TrimSpace(c.Request.URL.Path)
		if urlPath == urlPrefix {
			urlPath = path.Join(urlPrefix, indexHtml)
		}
		urlPath = path.Join(fsBase, urlPath)

		f, err := fs.Open(urlPath)
		if err != nil {
			return
		}
		fi, err := f.Stat()
		if strings.HasSuffix(urlPath, ".html") {
			c.Header("Cache-Control", "no-cache")
			c.Header("Content-Type", "text/html; charset=utf-8")
		}

		if strings.HasSuffix(urlPath, ".js") {
			c.Header("Content-Type", "text/javascript; charset=utf-8")
		}
		if strings.HasSuffix(urlPath, ".css") {
			c.Header("Content-Type", "text/css; charset=utf-8")
		}

		if err != nil || !fi.IsDir() {
			bs, err := fs.ReadFile(urlPath)
			if err != nil {
				logrus.WithError(err).Error("embed fs")
				return
			}
			c.Status(200)
			c.Writer.Write(bs)
			c.Abort()
		}
	}
}
