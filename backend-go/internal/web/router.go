package web

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// 公开：静态资源 + 订阅端点 + 登录/登出
	registerStatic(r)
	r.GET("/configs", h.queryConfig)
	r.POST("/login", h.login)
	r.POST("/logout", h.logout)

	// 受保护：业务 API
	api := r.Group("")
	api.Use(requireAuth(h.sessions))
	{
		api.GET("/clash_configs", h.listClashConfigs)
		api.POST("/clash_configs", h.createClashConfig)
		api.GET("/clash_configs/:id", h.detailClashConfig)
		api.PUT("/clash_configs/:id", h.updateClashConfig)
		api.DELETE("/clash_configs/:id", h.deleteClashConfig)

		api.GET("/clash_configs_merge", h.listMerges)
		api.POST("/clash_configs_merge", h.createMerge)
		api.GET("/clash_configs_merge/:id", h.detailMerge)
		api.PUT("/clash_configs_merge/:id", h.updateMerge)
		api.PUT("/clash_configs_merge/:id/token", h.refreshMergeToken)

		api.PUT("/users/me/password", h.changePassword)
	}

	return r
}
