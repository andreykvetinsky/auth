package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/andreykvetinsky/auth/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ServerConfig struct {
	ListenAddr string
}

type Server struct {
	config *config.Config
	router *gin.Engine

	authService authService
	logger      *zap.Logger
}

func NewServer(
	config *config.Config,

	authService authService,
	logger *zap.Logger,
) *Server {
	return &Server{
		config: config,
		router: gin.New(),

		authService: authService,
		logger:      logger,
	}
}

func (s *Server) Serve() error {
	s.setLogger()
	s.setRoutes()
	s.setTmeout()

	return s.run()
}

func (s *Server) setLogger() {
	s.router.Use(gin.Logger())
}

func (s *Server) setTmeout() {
	s.router.Use(s.timeoutMiddleware())
}

func (s *Server) setRoutes() {
	s.router.POST("/get-tokens", s.getTokens())
	s.router.POST("/refresh-tokens", s.refreshTokens())
}

func (s *Server) run() error {
	if err := s.router.Run(":" + s.config.HTTPServer.ListenAddr); err != nil {
		return fmt.Errorf("run server error :%w", err)
	}

	return nil
}

func (s *Server) timeoutMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), s.config.HTTPServer.RequestTimeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{}, 1)
		go func() {
			c.Next()
			done <- struct{}{}
		}()

		select {
		case <-ctx.Done():
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{"message": "request timed out"})
		case <-done:
		}
	}
}

func (s *Server) refreshTokens() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, exist := c.GetQuery("accessToken")
		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{"error": "not found accessToken in params"})

			return
		}

		refreshToken, exist := c.GetQuery("refreshToken")
		if !exist {
			c.JSON(http.StatusBadRequest, gin.H{"error": "not found refreshToken in params"})

			return
		}

		response, err := s.authService.RefereshTokens(c, accessToken, refreshToken)
		if err != nil {
			s.logger.Error("refresh tokens", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func (s *Server) getTokens() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.GetQuery("guid")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "not found guid in params"})

			return
		}

		response, err := s.authService.GetTokens(c, userID)
		if err != nil {
			s.logger.Error("get tokens", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

			return
		}

		c.JSON(http.StatusOK, response)
	}
}
