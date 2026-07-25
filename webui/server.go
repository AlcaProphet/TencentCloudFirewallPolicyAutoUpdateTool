package webui

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/alcaprophet/fwalizer/config"
)

// Server WebUI HTTP 服务器
type Server struct {
	store      *config.Store
	port       int
	mux        *http.ServeMux
	reloadFunc func() // 配置变更后通知 Syncer 重载
}

// NewServer 创建 WebUI 服务器
func NewServer(store *config.Store, port int) *Server {
	s := &Server{
		store: store,
		port:  port,
		mux:   http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// SetReloadFunc 设置配置重载回调（WebUI 修改配置后通知 Syncer）
func (s *Server) SetReloadFunc(fn func()) {
	s.reloadFunc = fn
}

// notifyReload 触发配置重载
func (s *Server) notifyReload() {
	if s.reloadFunc != nil {
		s.reloadFunc()
	}
}

// Start 启动 HTTP 服务器（阻塞）
func (s *Server) Start() error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	slog.Info("WebUI 启动", "addr", addr)
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) registerRoutes() {
	// 健康检查
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	// 目标管理
	s.mux.HandleFunc("GET /api/targets", s.handleGetTargets)
	s.mux.HandleFunc("POST /api/targets", s.handleAddTarget)
	s.mux.HandleFunc("DELETE /api/targets/{id}", s.handleDeleteTarget)
	// 规则管理
	s.mux.HandleFunc("GET /api/rules", s.handleGetRules)
	s.mux.HandleFunc("POST /api/rules", s.handleAddRule)
	s.mux.HandleFunc("DELETE /api/rules/{id}", s.handleDeleteRule)
	// 设置
	s.mux.HandleFunc("GET /api/settings", s.handleGetSettings)
	s.mux.HandleFunc("PUT /api/settings", s.handlePutSettings)
	// 同步日志
	s.mux.HandleFunc("GET /api/sync/logs", s.handleGetSyncLogs)
	// 静态文件（Vue SPA）
	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err == nil {
		s.mux.Handle("/", http.FileServer(http.FS(distFS)))
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := s.store.GetTargets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, targets)
}

func (s *Server) handleAddTarget(w http.ResponseWriter, r *http.Request) {
	var t config.TargetConfig
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.store.AddTarget(t); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.notifyReload()
	writeJSON(w, http.StatusCreated, map[string]string{"message": "添加成功"})
}

func (s *Server) handleDeleteTarget(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var n int
	fmt.Sscanf(id, "%d", &n)
	if err := s.store.DeleteTarget(n); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.notifyReload()
	writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

func (s *Server) handleGetRules(w http.ResponseWriter, r *http.Request) {
	rules, err := s.store.GetRules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (s *Server) handleAddRule(w http.ResponseWriter, r *http.Request) {
	var rule config.DomainRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.store.AddRule(rule); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.notifyReload()
	writeJSON(w, http.StatusCreated, map[string]string{"message": "添加成功"})
}

func (s *Server) handleDeleteRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var n int
	fmt.Sscanf(id, "%d", &n)
	if err := s.store.DeleteRule(n); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.notifyReload()
	writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.GetSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	var settings map[string]string
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	for k, v := range settings {
		if err := s.store.SetSetting(k, v); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	s.notifyReload()
	writeJSON(w, http.StatusOK, map[string]string{"message": "保存成功"})
}

func (s *Server) handleGetSyncLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := s.store.GetSyncLogs(100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError 写入错误响应
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
