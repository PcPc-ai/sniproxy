package proxy

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"sniproxy/internal/logger"
)

// handleHTTPConnection 处理HTTP连接
func (s *Server) handleHTTPConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// 设置读取超时
	clientConn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 读取HTTP请求
	reader := bufio.NewReader(clientConn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		logger.ErrorWithFields("读取HTTP请求失败", map[string]interface{}{
			"error":  err,
			"client": clientConn.RemoteAddr().String(),
		})
		return
	}

	// 获取主机名
	hostname := req.Host
	if hostname == "" {
		logger.ErrorWithFields("HTTP请求缺少Host头", map[string]interface{}{
			"client": clientConn.RemoteAddr().String(),
		})
		return
	}

	// 移除端口号（如果有）
	if strings.Contains(hostname, ":") {
		hostname = strings.Split(hostname, ":")[0]
	}

	logger.InfoWithFields("HTTP请求", map[string]interface{}{
		"host":   hostname,
		"method": req.Method,
		"uri":    req.RequestURI,
		"client": clientConn.RemoteAddr().String(),
	})

	// 检查白名单
	if !s.whitelist.IsAllowed(hostname) {
		logger.WarnWithFields("域名不在白名单中，连接被拒绝", map[string]interface{}{
			"hostname": hostname,
			"client":   clientConn.RemoteAddr().String(),
		})
		s.sendHTTPError(clientConn, 403, "Forbidden")
		return
	}

	// 确定目标端口
	targetPort := "80"
	if req.URL.Scheme == "https" {
		targetPort = "443"
	}

	// 如果原始Host包含端口，使用原始端口
	if strings.Contains(req.Host, ":") {
		_, port, err := net.SplitHostPort(req.Host)
		if err == nil {
			targetPort = port
		}
	}

	// 连接到目标服务器
	targetConn, err := s.connectToTarget(hostname, targetPort)
	if err != nil {
		logger.ErrorWithFields("连接目标服务器失败", map[string]interface{}{
			"hostname": hostname,
			"port":     targetPort,
			"error":    err,
			"client":   clientConn.RemoteAddr().String(),
		})
		s.sendHTTPError(clientConn, 502, "Bad Gateway")
		return
	}
	defer targetConn.Close()

	// 重构请求以转发
	requestLine := fmt.Sprintf("%s %s %s\r\n", req.Method, req.URL.RequestURI(), req.Proto)

	// 写入请求行
	_, err = targetConn.Write([]byte(requestLine))
	if err != nil {
		logger.ErrorWithFields("转发请求行失败", map[string]interface{}{
			"error":    err,
			"hostname": hostname,
			"client":   clientConn.RemoteAddr().String(),
		})
		return
	}

	// 写入请求头
	for key, values := range req.Header {
		for _, value := range values {
			header := fmt.Sprintf("%s: %s\r\n", key, value)
			_, err = targetConn.Write([]byte(header))
			if err != nil {
				logger.ErrorWithFields("转发请求头失败", map[string]interface{}{
					"error":    err,
					"header":   key,
					"hostname": hostname,
					"client":   clientConn.RemoteAddr().String(),
				})
				return
			}
		}
	}

	// 写入空行结束头部
	_, err = targetConn.Write([]byte("\r\n"))
	if err != nil {
		logger.ErrorWithFields("转发请求结束失败", map[string]interface{}{
			"error":    err,
			"hostname": hostname,
			"client":   clientConn.RemoteAddr().String(),
		})
		return
	}

	// 如果有请求体，转发请求体
	if req.ContentLength > 0 {
		_, err = targetConn.Write(make([]byte, req.ContentLength))
		if err != nil {
			logger.ErrorWithFields("转发请求体失败", map[string]interface{}{
				"error":    err,
				"hostname": hostname,
				"client":   clientConn.RemoteAddr().String(),
			})
			return
		}
	}

	// 清除读取超时
	clientConn.SetReadDeadline(time.Time{})

	// 开始双向数据转发
	logger.InfoWithFields("开始代理HTTP连接", map[string]interface{}{
		"client": clientConn.RemoteAddr().String(),
		"target": hostname + ":" + targetPort,
	})
	s.relay(clientConn, targetConn)
	logger.DebugWithFields("HTTP连接结束", map[string]interface{}{
		"client": clientConn.RemoteAddr().String(),
		"target": hostname + ":" + targetPort,
	})
}

// sendHTTPError 发送HTTP错误响应
func (s *Server) sendHTTPError(conn net.Conn, code int, message string) {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: text/plain\r\nConnection: close\r\n\r\n%d %s\n",
		code, message, code, message)
	conn.Write([]byte(response))
}
