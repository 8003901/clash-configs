package web

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/8003901/clash-configs/backend-go/internal/merge"
	"github.com/8003901/clash-configs/backend-go/internal/model"
	"github.com/8003901/clash-configs/backend-go/internal/service"
	"github.com/8003901/clash-configs/backend-go/internal/store"
)

type Handler struct {
	store    *store.Store
	cc       *service.ClashConfigService
	ms       *service.MergeService
	us       *service.UserService
	sessions *SessionStore
}

func NewHandler(s *store.Store, cc *service.ClashConfigService, ms *service.MergeService, us *service.UserService) *Handler {
	return &Handler{store: s, cc: cc, ms: ms, us: us, sessions: NewSessionStore()}
}

// --- auth ---

func (h *Handler) login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if !h.us.Authenticate(username, password) {
		c.Redirect(http.StatusFound, "/login?error")
		return
	}
	id := h.sessions.Create(username)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookie, id, 0, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func (h *Handler) logout(c *gin.Context) {
	if id, err := c.Cookie(sessionCookie); err == nil {
		h.sessions.Delete(id)
	}
	c.SetCookie(sessionCookie, "", -1, "/", "", false, true)
	c.Status(http.StatusOK)
}

func (h *Handler) changePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	username, ok := h.currentUsername(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}
	if err := h.us.ChangePwd(username, req.NewPassword, req.OldPassword); err != nil {
		if errors.Is(err, service.ErrBadOldPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "旧密码错误"})
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) currentUsername(c *gin.Context) (string, bool) {
	id, err := c.Cookie(sessionCookie)
	if err != nil {
		return "", false
	}
	return h.sessions.Get(id)
}

// --- clash_configs ---

type clashConfigAdd struct {
	URL            string `json:"url"`
	Name           string `json:"name"`
	UpdateSchedule string `json:"updateSchedule"`
	Enabled        bool   `json:"enabled"`
}

func (h *Handler) createClashConfig(c *gin.Context) {
	var req clashConfigAdd
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if req.URL == "" || req.Name == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	if req.UpdateSchedule == "" {
		req.UpdateSchedule = "DAY"
	}
	if req.UpdateSchedule != "DAY" && req.UpdateSchedule != "WEEK" {
		c.Status(http.StatusBadRequest)
		return
	}
	cc := &model.ClashConfig{URL: req.URL, Name: req.Name, Enabled: req.Enabled, UpdateSchedule: req.UpdateSchedule}
	saved, err := h.cc.Save(cc)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (h *Handler) updateClashConfig(c *gin.Context) {
	var req clashConfigAdd
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if req.URL == "" || req.Name == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	cc := &model.ClashConfig{ID: c.Param("id"), URL: req.URL, Name: req.Name, Enabled: req.Enabled, UpdateSchedule: req.UpdateSchedule}
	saved, err := h.cc.Save(cc)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (h *Handler) detailClashConfig(c *gin.Context) {
	renew := c.Query("renew") == "true"
	cc, err := h.cc.Detail(c.Param("id"), renew)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, cc)
}

func (h *Handler) listClashConfigs(c *gin.Context) {
	list, err := h.cc.All()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) deleteClashConfig(c *gin.Context) {
	if err := h.cc.Delete(c.Param("id")); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

// --- clash_configs_merge ---

type mergeAdd struct {
	Name      string   `json:"name"`
	ConfigIDs []string `json:"configIds"`
}

func (h *Handler) createMerge(c *gin.Context) {
	var req mergeAdd
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	m, err := h.ms.Create(req.Name, req.ConfigIDs)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) listMerges(c *gin.Context) {
	list, err := h.ms.List()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) detailMerge(c *gin.Context) {
	m, err := h.ms.Detail(c.Param("id"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, m)
}

type mergeUpdate struct {
	Name    string              `json:"name"`
	Config  string              `json:"config"`
	Configs []model.ClashConfig `json:"configs"`
}

func (h *Handler) updateMerge(c *gin.Context) {
	var req mergeUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Config == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	m := &model.ClashConfigsMerge{ID: c.Param("id"), Name: req.Name, Config: req.Config, Configs: req.Configs}
	saved, err := h.ms.Save(m)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (h *Handler) refreshMergeToken(c *gin.Context) {
	m, err := h.ms.RefreshToken(c.Param("id"))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, m)
}

// --- /configs ---

func (h *Handler) queryConfig(c *gin.Context) {
	token := c.Query("token")
	m, err := h.ms.FindByToken(token)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
		}
		return
	}
	sources := make([]merge.Source, 0, len(m.Configs))
	for _, cc := range m.Configs {
		sources = append(sources, merge.Source{Content: cc.Content, UserInfo: cc.SubscriptionUserinfo})
	}
	res, err := merge.Merge(m.Config, sources)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	if res.UserInfo != nil {
		c.Header("subscription-userinfo", res.UserInfo.HeaderValue())
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(res.YAML))
}
