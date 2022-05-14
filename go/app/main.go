package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"encoding/json"
	"io/ioutil"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "image"
)

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

type Item struct {
	Items []Contents `json:"items"`
}

type Contents struct {
	Name string `json:"name"`
	Category string `json:"category"`
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Receive item: %s %s", name, category)

	bytes, err := ioutil.ReadFile("app/items.json")
	if err != nil {
		log.Fatal(err)
	}

	var item Item
	if err := json.Unmarshal(bytes, &item); err != nil {
        log.Fatal(err)
    }

	var contents Contents
	contents.Name = name
	contents.Category = category
	item.Items = append(item.Items, contents)

	n_json, err := json.Marshal(item)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile("app/items.json", n_json, os.ModePerm)

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func showItem(c echo.Context) error {
	bytes, err := ioutil.ReadFile("app/items.json")
	if err != nil {
		log.Fatal(err)
	}

	message := string(bytes) + "\n"
	return c.String(http.StatusOK, message)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("itemImg"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", showItem)
	e.GET("/image/:itemImg", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
