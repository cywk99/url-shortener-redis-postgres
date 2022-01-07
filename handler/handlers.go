package handler

import (
	"fmt"
	"net/http"

	"github.com/cywk/go-url-shortener/shortener"
	store "github.com/cywk/go-url-shortener/storage"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
)

// Request model definition
type UrlCreationRequest struct {
	LongUrl string `json:"long_url" binding:"required"`
}

func CreateShortUrl(c *gin.Context) {
	var creationRequest UrlCreationRequest
	if err := c.ShouldBindJSON(&creationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user_id := uniuri.New()
	shortUrl := shortener.GenerateShortLink(creationRequest.LongUrl, user_id)
	store.SaveUrlMapping(shortUrl, creationRequest.LongUrl, user_id)

	host := "http://localhost:9808/"
	c.JSON(200, gin.H{
		"message":   "short url created successfully",
		"short_url": host + shortUrl,
	})

}

func HandleShortUrlRedirect(c *gin.Context) {
	shortUrl := c.Param("shortUrl")
	fmt.Println(shortUrl)
	initialUrl := store.RetrieveInitialUrl(shortUrl)
	c.Redirect(302, initialUrl)
}
