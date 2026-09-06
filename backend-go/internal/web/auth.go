package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// isAjax 判断是否为前端 AJAX 请求（对齐 Spring Security 的判定逻辑）。
func isAjax(r *http.Request) bool {
	if strings.EqualFold(r.Header.Get("X-Requested-With"), "XMLHttpRequest") {
		return true
	}
	return strings.Contains(r.Header.Get("Accept"), "application/json")
}

// requireAuth 鉴权中间件：未认证时 AJAX 返回 401 JSON，否则 302 跳转 /login。
// 独立函数（不依赖 Handler），由 Task 6 的路由以 requireAuth(h.sessions) 挂载。
func requireAuth(sessions *SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := c.Cookie(sessionCookie)
		if err == nil {
			if _, ok := sessions.Get(id); ok {
				c.Next()
				return
			}
		}
		if isAjax(c.Request) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
	}
}
