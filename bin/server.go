package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vincent-petithory/dataurl"
)

func OCRHandler(c echo.Context) error {
	image := c.FormValue("image")
	lang := c.FormValue("lang")

	dataURL, err := dataurl.DecodeString(image)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	var file *os.File
	switch dataURL.ContentType() {
	case "image/png":
		file, err = ioutil.TempFile("", "*.png")
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		file.Write(dataURL.Data)
	case "image/jpeg":
		file, err = ioutil.TempFile("", "*.jpeg")
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		file.Write(dataURL.Data)
	default:
		return c.String(
			http.StatusBadRequest,
			fmt.Errorf("only png and jpeg images are supported").Error(),
		)
	}
	defer os.Remove(file.Name())
	file.Close()

	outputFile, err := ioutil.TempFile("", "*")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer os.Remove(outputFile.Name())
	outputFile.Close()

	command := exec.Command("tesseract", file.Name(), outputFile.Name(), "-l", lang)

	output, err := command.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "Tesseract couldn't load any languages!") {
			return c.String(http.StatusInternalServerError, fmt.Errorf("invalid language(s)").Error())
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}
	parsed, err := ioutil.ReadFile(outputFile.Name() + ".txt")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, string(parsed))
}

func main() {
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.BodyLimit("4M"))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper: middleware.DefaultSkipper,
		Timeout: time.Second * 10,
	}))

	e.POST("/", OCRHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
