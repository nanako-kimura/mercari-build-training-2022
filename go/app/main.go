package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"encoding/json"
	"database/sql"
	"encoding/hex"
	"crypto/sha256"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	_ "github.com/mattn/go-sqlite3"
)

var DbConnection *sql.DB

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
	Image string `json:"image"`
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	i_path := c.FormValue("image")
	hashed:=sha256.Sum256([]byte(i_path))
	image := hex.EncodeToString(hashed[:])+".jpg"

	c.Logger().Infof("Receive item: %s %s %s", name, category, image)

	DbConnection,_:=sql.Open("sqlite3","../db/mercari.sqlite3")
	defer DbConnection.Close()

	cmd := "INSERT INTO items (name,category,image) VALUES (?,?,?)"
	_, err := DbConnection.Exec(cmd,name,category,image)
	if err != nil {
		log.Fatal(err)
	}

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func showItem(c echo.Context) error {
	DbConnection,_:=sql.Open("sqlite3","../db/mercari.sqlite3")
	defer DbConnection.Close()

	cmd := "SELECT name,category,image FROM items"
	rows, err := DbConnection.Query(cmd)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	var item Item
	for rows.Next(){
		var contents Contents
		err := rows.Scan(&contents.Name,&contents.Category,&contents.Image)
		if err != nil{
			log.Fatal(err)
		}
		item.Items = append(item.Items,contents)
	}
	n_json, err := json.Marshal(item)
	if err != nil {
		log.Fatal(err)
	}
	return c.String(http.StatusOK, string(n_json)+"\n")
}

func searchItem(c echo.Context) error {
	name := c.FormValue("keyword")

	DbConnection,_:=sql.Open("sqlite3","../db/mercari.sqlite3")
	defer DbConnection.Close()

	cmd := "SELECT name,category,image FROM items where name = ?"
	rows, err := DbConnection.Query(cmd,name)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var item Item
	for rows.Next(){
		var contents Contents
		err := rows.Scan(&contents.Name,&contents.Category,&contents.Image)
		if err != nil{
			log.Fatal(err)
		}
		item.Items = append(item.Items,contents)
	}

	n_json, err := json.Marshal(item)
	if err != nil {
		log.Fatal(err)
	}
	return c.String(http.StatusOK, string(n_json)+"\n")
}

func getItem(c echo.Context) error {
	// Create image path
	id := c.Param("id")
	DbConnection,_:=sql.Open("sqlite3","../db/mercari.sqlite3")
	defer DbConnection.Close()

	cmd := "SELECT name,category,image FROM items where id = ?"
	row := DbConnection.QueryRow(cmd,id)

	var contents Contents
	err := row.Scan(&contents.Name,&contents.Category,&contents.Image)
	if err != nil{
		if err != sql.ErrNoRows{
			log.Fatal(err)
		}
	}
	n_json, err := json.Marshal(contents)
	if err != nil {
		log.Fatal(err)
	}
	return c.String(http.StatusOK, string(n_json)+"\n")
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
	e.GET("/items/:id", getItem)
	e.GET("/search", searchItem)
	e.GET("/image/:itemImg", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
