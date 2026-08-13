package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"
)

// selfSignedCert 生成自签证书（ECDSA P-256，10 年有效），SAN 包含指定域名/IP。
// 免去证书文件管理；浏览器首次访问需手动信任（https 即安全上下文，
// crypto.randomUUID 等 Web API 全部可用）
func selfSignedCert(hosts []string) (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: hosts[0]},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		} else {
			tmpl.DNSNames = append(tmpl.DNSNames, h)
		}
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, nil
}

// startTLS 启动自签 HTTPS 监听（与 HTTP 共用同一套 handler），阻塞直到退出
func startTLS(addr, hosts string) error {
	list := []string{}
	for _, h := range strings.Split(hosts, ",") {
		if h = strings.TrimSpace(h); h != "" {
			list = append(list, h)
		}
	}
	cert, err := selfSignedCert(list)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:      addr,
		Handler:   nil, // 与 HTTP 共用 DefaultServeMux
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
	}
	log.Printf("self-signed HTTPS on %s (hosts: %s) — 浏览器需手动信任证书", addr, hosts)
	return server.ListenAndServeTLS("", "")
}
