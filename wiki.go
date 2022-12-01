package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type Page struct {
	Title string
	Body  string
}

var basePath = os.Args[1]
var dataPath = basePath + "/data/"

func (p *Page) save() error {
	filename := strings.Join([]string{dataPath, p.Title, ".txt"}, "")
	return os.WriteFile(filename, []byte(p.Body), 0600)
}

func loadPage(title string) (*Page, error) {
	filename := strings.Join([]string{dataPath, title, ".txt"}, "")
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: string(body)}, nil
}

var innerLink = regexp.MustCompile(`\[([a-zA-Z0-9]+)\]`)

func (p *Page) Display() map[string]any {
	dp := map[string]any{"Title": p.Title}
	dp["Body"] = template.HTML(innerLink.ReplaceAllStringFunc(p.Body, func(match string) string {
		matchStr := match[1 : len(match)-1]
		return strings.Join([]string{"<a href=\"/view/", matchStr, "\">", matchStr, "</a>"}, "")
	}))
	return dp
}

func home(c *gin.Context) {
	c.Redirect(http.StatusFound, "/view/FrontPage")
}

func viewing(c *gin.Context) {
	title := c.Param("title")
	p, err := loadPage(title)
	if err != nil {
		c.Redirect(http.StatusFound, "/edit/"+title)
		return
	}
	c.HTML(http.StatusOK, "view.html", p.Display())
}

func editing(c *gin.Context) {
	title := c.Param("title")
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	c.HTML(http.StatusOK, "edit.html", p)
}

func saving(c *gin.Context) {
	title := c.Param("title")
	body := c.PostForm("body")
	p := &Page{Title: title, Body: body}
	err := p.save()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, "/view/"+title)
}

func main() {
	router := gin.Default()

	router.LoadHTMLGlob(basePath + "/templates/*.html")

	router.GET("/", home)
	router.GET("/view/:title", viewing)
	router.GET("/edit/:title", editing)
	router.POST("/save/:title", saving)
	router.Static("/static", basePath+"/static")

	err := router.Run(":8080")
	if err != nil {
		fmt.Println(err)
	}
}
