package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"

	"sniproxy/internal/config"
	"sniproxy/internal/logger"
	"sniproxy/internal/whitelist"
)

// Server SNI代理服务器
type Server struct {
	config        *config.Config
	whitelist     *whitelist.Whitelist
	httpsListener net.Listener
	httpListener  net.Listener
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewServer 创建新的代理服务器
func NewServer(cfg *config.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	server := &Server{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	// 初始化白名单
	server.whitelist = whitelist.NewWhitelist(
		cfg.Whitelist.Enabled,
		cfg.Whitelist.DefaultMode,
	)

	// 加载白名单规则
	if cfg.Whitelist.DomainsFile != "" {
		if err := server.whitelist.LoadFromFile(cfg.Whitelist.DomainsFile); err != nil {
			logger.Warn("加载白名单文件失败: %v", err)
		} else {
			logger.InfoWithFields("已加载白名单规则", map[string]interface{}{
				"file":  cfg.Whitelist.DomainsFile,
				"count": server.whitelist.GetRulesCount(),
			})
		}
	}

	// 从配置中添加域名
	if len(cfg.Whitelist.Domains) > 0 {
		server.whitelist.AddRules(cfg.Whitelist.Domains, "suffix")
		logger.InfoWithFields("从配置添加白名单规则", map[string]interface{}{
			"count": len(cfg.Whitelist.Domains),
		})
	}

	return server
}

// Start 启动代理服务器
func (s *Server) Start() error {
	// 启动HTTPS代理
	httpsAddr := fmt.Sprintf("%s:%s", s.config.Server.BindAddr, s.config.Server.HTTPSPort)
	httpsListener, err := net.Listen("tcp", httpsAddr)
	if err != nil {
		return fmt.Errorf("监听HTTPS端口失败: %w", err)
	}
	s.httpsListener = httpsListener
	logger.InfoWithFields("HTTPS代理已启动", map[string]interface{}{
		"address": httpsAddr,
	})

	// 启动HTTP代理
	httpAddr := fmt.Sprintf("%s:%s", s.config.Server.BindAddr, s.config.Server.HTTPPort)
	httpListener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		httpsListener.Close()
		return fmt.Errorf("监听HTTP端口失败: %w", err)
	}
	s.httpListener = httpListener
	logger.InfoWithFields("HTTP代理已启动", map[string]interface{}{
		"address": httpAddr,
	})

	// 启动HTTPS处理协程
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.handleHTTPS()
	}()

	// 启动HTTP处理协程
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.handleHTTP()
	}()

	return nil
}

// Stop 停止代理服务器
func (s *Server) Stop() {
	s.cancel()

	if s.httpsListener != nil {
		s.httpsListener.Close()
	}
	if s.httpListener != nil {
		s.httpListener.Close()
	}

	s.wg.Wait()
	logger.Info("代理服务器已停止")
}

// handleHTTPS 处理HTTPS连接
func (s *Server) handleHTTPS() {
	for {
		conn, err := s.httpsListener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				logger.Error("接受HTTPS连接失败: %v", err)
				continue
			}
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleHTTPSConnection(conn)
		}()
	}
}

// handleHTTP 处理HTTP连接
func (s *Server) handleHTTP() {
	for {
		conn, err := s.httpListener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				logger.Error("接受HTTP连接失败: %v", err)
				continue
			}
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleHTTPConnection(conn)
		}()
	}
}
