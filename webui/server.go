package webui

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/alcaprophet/fwalizer/config"
	"github.com/alcaprophet/fwalizer/webui/api"
)

// Server WebUI HTTP 服务器
type Server struct {
	store *config.Store
	port  int
	mux   *http.ServeMux
	deps  *api.Deps
}

// NewServer 创建 WebUI 服务器
func NewServer(store *config.Store, port int) *Server {
	s := &Server{
		store: store,
		port:  port,
		mux:   http.NewServeMux(),
		deps:  &api.Deps{Store: store},
	}
	s.registerRoutes()
	return s
}

// SetSyncer 设置同步引擎和事件总线（WebUI 模式下由 main.go 调用）
func (s *Server) SetSyncer(sync api.Syncer, bus api.EventSubscriber) {
	s.deps.Syncer = sync
	s.deps.EventBus = bus
}

// SetReloadFunc 设置配置重载回调（WebUI 修改配置后通知 Syncer）
func (s *Server) SetReloadFunc(fn func()) {
	s.deps.ReloadFunc = fn
}

// Start 启动 HTTP 服务器（阻塞）
func (s *Server) Start() error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	slog.Info("WebUI 启动", "addr", addr)
	return http.ListenAndServe(addr, s.mux)
}

// registerRoutes 注册路由：health + API 委托 + 静态文件
func (s *Server) registerRoutes() {
	// 健康检查
	s.mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"status":"ok"}`))
	})
	// 所有业务 API 端点
	s.deps.Register(s.mux)
	// 静态文件（Vue SPA）
	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err == nil {
		s.mux.Handle("/", http.FileServer(http.FS(distFS)))
	}
}
