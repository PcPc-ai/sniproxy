package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sniproxy/internal/config"
	"sniproxy/internal/logger"
	"sniproxy/internal/proxy"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化日志系统
	logger.InitLogger(cfg.Logging.Level, cfg.Logging.Format)
	logger.Info("SNI代理服务器启动")
	logger.InfoWithFields("已加载配置", map[string]interface{}{
		"config_file": configFile,
		"https_port":  cfg.Server.HTTPSPort,
		"http_port":   cfg.Server.HTTPPort,
		"bind_addr":   cfg.Server.BindAddr,
	})

	// 创建代理服务器
	server := proxy.NewServer(cfg)

	// 启动服务器
	if err := server.Start(); err != nil {
		logger.Error("启动服务器失败: %v", err)
		log.Fatalf("启动服务器失败: %v", err)
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("收到停止信号，正在关闭服务器...")
	server.Stop()
	logger.Info("服务器已关闭")
}
