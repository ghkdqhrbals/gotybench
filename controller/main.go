package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")

	// 입력 폼을 렌더링하는 핸들러
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	// 입력 데이터를 처리하는 핸들러
	router.POST("/submit", func(c *gin.Context) {
		input1 := c.PostForm("input1")
		input2 := c.PostForm("input2")
		// 받은 입력 데이터를 처리하는 로직을 추가
		c.String(200, "Input 1: %s\nInput 2: %s\n", input1, input2)
	})

	// 웹 서버 시작
	router.Run(":8020")
}
