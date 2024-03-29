package controller

import (
	"github.com/gin-gonic/gin"
	"goPandora/model"
	"net/http"
)

func error404(c *gin.Context) {
	props := gin.H{
		"props": gin.H{
			"pageProps": gin.H{"statusCode": 404},
		},
		"page":         "/_error",
		"query":        gin.H{},
		"buildId":      model.Param.BuildId,
		"assetPrefix":  nil,
		"nextExport":   true,
		"isFallback":   false,
		"gip":          true,
		"scriptLoader": "[]",
	}
	c.HTML(http.StatusNotFound, "404.html", gin.H{
		"props":      props,
		"api_prefix": model.Param.ApiPrefix,
	})
}
