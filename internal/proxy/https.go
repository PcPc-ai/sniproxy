package proxy

import (
	"fmt"
	"io"
	"net"
	"time"

	"sniproxy/internal/logger"
	"sniproxy/internal/sni"
)

// handleHTTPSConnection 处理HTTPS连接
func (s *Server) handleHTTPSConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// 设置读取超时
	clientConn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 读取ClientHello消息
	buffer := make([]byte, 4096)
	n, err := clientConn.Read(buffer)
	if err != nil {
		logger.ErrorWithFields("读取ClientHello失败", map[string]interface{}{
			"error":  err,
			"client": clientConn.RemoteAddr().String(),
		})
		return
	}

	// 解析SNI
	hostname, err := sni.ParseSNI(buffer[:n])
	if err != nil {
		// logger.ErrorWithFields("解析SNI失败", map[string]interface{}{
		// 	"error":  err,
		// 	"client": clientConn.RemoteAddr().String(),
		// })
		return
	}

	logger.InfoWithFields("HTTPS请求", map[string]interface{}{
		"sni":    hostname,
		"client": clientConn.RemoteAddr().String(),
	})

	// 检查白名单
	if !s.whitelist.IsAllowed(hostname) {
		logger.WarnWithFields("域名不在白名单中，连接被拒绝", map[string]interface{}{
			"hostname": hostname,
			"client":   clientConn.RemoteAddr().String(),
		})
		return
	}

	// 连接到目标服务器
	targetConn, err := s.connectToTarget(hostname, "443")
	if err != nil {
		logger.ErrorWithFields("连接目标服务器失败", map[string]interface{}{
			"hostname": hostname,
			"port":     "443",
			"error":    err,
			"client":   clientConn.RemoteAddr().String(),
		})
		return
	}
	defer targetConn.Close()

	// 将ClientHello转发给目标服务器
	_, err = targetConn.Write(buffer[:n])
	if err != nil {
		logger.ErrorWithFields("转发ClientHello失败", map[string]interface{}{
			"error":    err,
			"hostname": hostname,
			"client":   clientConn.RemoteAddr().String(),
		})
		return
	}

	// 清除读取超时
	clientConn.SetReadDeadline(time.Time{})

	// // 开始双向数据转发
	// logger.InfoWithFields("开始代理HTTPS连接", map[string]interface{}{
	// 	"client": clientConn.RemoteAddr().String(),
	// 	"target": hostname + ":443",
	// })
	s.relay(clientConn, targetConn)
	// logger.DebugWithFields("HTTPS连接结束", map[string]interface{}{
	// 	"client": clientConn.RemoteAddr().String(),
	// 	"target": hostname + ":443",
	// })
}

// connectToTarget 连接到目标服务器
func (s *Server) connectToTarget(hostname, port string) (net.Conn, error) {
	target := net.JoinHostPort(hostname, port)

	// 设置连接超时
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := dialer.Dial("tcp", target)
	if err != nil {
		return nil, fmt.Errorf("连接 %s 失败: %w", target, err)
	}

	return conn, nil
}

// relay 双向数据转发
func (s *Server) relay(client, target net.Conn) {
	done := make(chan struct{}, 2)

	// 客户端到服务器
	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(target, client)
		target.Close()
	}()

	// 服务器到客户端
	go func() {
		defer func() { done <- struct{}{} }()
		io.Copy(client, target)
		client.Close()
	}()

	// 等待任一方向完成
	<-done
}
