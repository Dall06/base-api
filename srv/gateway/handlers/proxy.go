package handlers

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"base-api/pkg/errs"

	"github.com/labstack/echo/v4"
)

type ProxyHandler struct {
	userURL string
	client  *http.Client
}

func NewProxyHandler(userURL string) *ProxyHandler {
	return &ProxyHandler{
		userURL: strings.TrimSuffix(userURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				MaxConnsPerHost:     50,
				IdleConnTimeout:     90 * time.Second,
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
			},
		},
	}
}

func (h *ProxyHandler) ProxyToUser(ctx echo.Context) error {
	targetPath := ctx.Request().URL.Path
	targetURL := h.userURL + targetPath

	if ctx.Request().URL.RawQuery != "" {
		targetURL += "?" + ctx.Request().URL.RawQuery
	}

	return h.proxyRequest(ctx, targetURL)
}

func (h *ProxyHandler) proxyRequest(ctx echo.Context, targetURL string) error {
	req := ctx.Request()

	var body io.Reader
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return errs.Handle(ctx, errs.ValueError("no se pudo leer el cuerpo de la petición: %v", err))
		}
		body = bytes.NewReader(bodyBytes)
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	proxyReq, err := http.NewRequest(req.Method, targetURL, body)
	if err != nil {
		return errs.Handle(ctx, errs.InternalError("error al crear petición proxy: %v", err))
	}

	// Copiar cabeceras
	for key, values := range req.Header {
		for _, val := range values {
			proxyReq.Header.Add(key, val)
		}
	}

	// Cabeceras proxy estándar
	proxyReq.Header.Set("X-Forwarded-For", req.RemoteAddr)

	resp, err := h.client.Do(proxyReq)
	if err != nil {
		slog.Error("error al enviar petición proxy", slog.String("url", targetURL), slog.String("error", err.Error()))
		return errs.Handle(ctx, errs.ServiceUnavailableError("servicio temporalmente no disponible"))
	}
	defer resp.Body.Close()

	// Copiar cabeceras de respuesta
	for key, values := range resp.Header {
		for _, val := range values {
			ctx.Response().Header().Add(key, val)
		}
	}

	ctx.Response().WriteHeader(resp.StatusCode)

	_, err = io.Copy(ctx.Response().Writer, resp.Body)
	return err
}
