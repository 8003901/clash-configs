package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed static
var staticFS embed.FS

// staticSub 返回以 static/ 为根的文件系统（含 index.html、assets/、favicon.svg、icons.svg）。
func staticSub() http.FileSystem {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

func registerStatic(r *gin.Engine) {
	// http.FileServer 把 URL 路径映射到 static/ 下的文件：
	//   /              -> static/index.html（目录默认返回 index.html）
	//   /assets/x.js   -> static/assets/x.js
	//   /favicon.svg   -> static/favicon.svg
	fileServer := gin.WrapH(http.FileServer(staticSub()))
	r.GET("/", fileServer)
	r.GET("/index.html", fileServer)
	r.GET("/assets/*filepath", fileServer)
	r.GET("/favicon.svg", fileServer)
	r.GET("/icons.svg", fileServer)
}
