package sni

import (
	"errors"
)

// TLS记录类型
const (
	RecordTypeHandshake = 0x16
)

// TLS握手类型
const (
	HandshakeTypeClientHello = 0x01
)

// ParseSNI 从TLS ClientHello消息中解析SNI
func ParseSNI(data []byte) (string, error) {
	if len(data) < 5 {
		return "", errors.New("数据太短，不是有效的TLS记录")
	}

	// 检查TLS记录头
	if data[0] != RecordTypeHandshake {
		return "", errors.New("不是TLS握手记录")
	}

	// 跳过TLS记录头 (5字节)
	offset := 5

	if offset >= len(data) {
		return "", errors.New("TLS记录数据不完整")
	}

	// 检查握手类型
	if data[offset] != HandshakeTypeClientHello {
		return "", errors.New("不是ClientHello消息")
	}

	// 跳过握手消息头 (4字节: 类型(1) + 长度(3))
	offset += 4

	// 跳过协议版本 (2字节)
	offset += 2

	if offset+32 > len(data) {
		return "", errors.New("ClientHello数据不完整")
	}

	// 跳过随机数 (32字节)
	offset += 32

	if offset >= len(data) {
		return "", errors.New("ClientHello数据不完整")
	}

	// 跳过会话ID
	sessionIDLength := int(data[offset])
	offset += 1 + sessionIDLength

	if offset+2 > len(data) {
		return "", errors.New("ClientHello数据不完整")
	}

	// 跳过密码套件
	cipherSuitesLength := int(data[offset])<<8 | int(data[offset+1])
	offset += 2 + cipherSuitesLength

	if offset >= len(data) {
		return "", errors.New("ClientHello数据不完整")
	}

	// 跳过压缩方法
	compressionMethodsLength := int(data[offset])
	offset += 1 + compressionMethodsLength

	if offset+2 > len(data) {
		return "", errors.New("没有扩展数据")
	}

	// 解析扩展
	extensionsLength := int(data[offset])<<8 | int(data[offset+1])
	offset += 2

	if offset+extensionsLength > len(data) {
		return "", errors.New("扩展数据不完整")
	}

	return parseExtensions(data[offset : offset+extensionsLength])
}

// parseExtensions 解析TLS扩展，查找SNI扩展
func parseExtensions(extensions []byte) (string, error) {
	offset := 0

	for offset+4 <= len(extensions) {
		// 读取扩展类型和长度
		extType := int(extensions[offset])<<8 | int(extensions[offset+1])
		extLength := int(extensions[offset+2])<<8 | int(extensions[offset+3])
		offset += 4

		if offset+extLength > len(extensions) {
			return "", errors.New("扩展数据长度错误")
		}

		// SNI扩展类型是0x0000
		if extType == 0x0000 {
			return parseSNIExtension(extensions[offset : offset+extLength])
		}

		offset += extLength
	}

	return "", errors.New("未找到SNI扩展")
}

// parseSNIExtension 解析SNI扩展
func parseSNIExtension(sniData []byte) (string, error) {
	if len(sniData) < 2 {
		return "", errors.New("SNI扩展数据太短")
	}

	// 读取服务器名称列表长度
	listLength := int(sniData[0])<<8 | int(sniData[1])
	offset := 2

	if offset+listLength > len(sniData) {
		return "", errors.New("SNI服务器名称列表长度错误")
	}

	for offset < 2+listLength {
		if offset+3 > len(sniData) {
			return "", errors.New("SNI名称条目数据不完整")
		}

		// 读取名称类型和长度
		nameType := sniData[offset]
		nameLength := int(sniData[offset+1])<<8 | int(sniData[offset+2])
		offset += 3

		if offset+nameLength > len(sniData) {
			return "", errors.New("SNI名称数据不完整")
		}

		// 名称类型0表示主机名
		if nameType == 0 {
			hostname := string(sniData[offset : offset+nameLength])
			return hostname, nil
		}

		offset += nameLength
	}

	return "", errors.New("未找到主机名")
}

// IsValidSNI 检查是否为有效的SNI数据
func IsValidSNI(data []byte) bool {
	_, err := ParseSNI(data)
	return err == nil
}

// PeekSNI 尝试从数据中提取SNI，不返回错误
func PeekSNI(data []byte) string {
	sni, err := ParseSNI(data)
	if err != nil {
		return ""
	}
	return sni
}
