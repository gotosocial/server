/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package router

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/sirupsen/logrus"
)

// Router provides the REST interface for gotosocial, using gin.
type Router interface {
	// Attach a gin handler to the router with the given method and path
	AttachHandler(method string, path string, handler gin.HandlerFunc)
	// Attach a gin middleware to the router that will be used globally
	AttachMiddleware(handler gin.HandlerFunc)
	// Start the router
	Start()
	// Stop the router
	Stop()
}

// router fulfils the Router interface using gin and logrus
type router struct {
	logger *logrus.Logger
	engine *gin.Engine
}

// Start starts the router nicely
func (s *router) Start() {
	// todo: start gracefully
	if err := s.engine.Run(); err != nil {
		s.logger.Panicf("server error: %s", err)
	}
}

// Stop shuts down the router nicely
func (s *router) Stop() {
	// todo: shut down gracefully
}

// AttachHandler attaches the given gin.HandlerFunc to the router with the specified method and path.
// If the path is set to ANY, then the handlerfunc will be used for ALL methods at its given path.
func (s *router) AttachHandler(method string, path string, handler gin.HandlerFunc) {
	if method == "ANY" {
		s.engine.Any(path, handler)
	} else {
		s.engine.Handle(method, path, handler)
	}
}

// AttachMiddleware attaches a gin middleware to the router that will be used globally
func (s *router) AttachMiddleware(middleware gin.HandlerFunc) {
	s.engine.Use(middleware)
}

// New returns a new Router with the specified configuration, using the given logrus logger.
func New(config *config.Config, logger *logrus.Logger) (Router, error) {
	engine := gin.New()

	// create a new session store middleware
	store, err := sessionStore()
	if err != nil {
		return nil, fmt.Errorf("error creating session store: %s", err)
	}
	engine.Use(sessions.Sessions("gotosocial-session", store))

	// load html templates for use by the router
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory: %s", err)
	}
	tmPath := filepath.Join(cwd, fmt.Sprintf("%s*", config.TemplateConfig.BaseDir))
	logger.Debugf("loading templates from %s", tmPath)
	engine.LoadHTMLGlob(tmPath)

	return &router{
		logger: logger,
		engine: engine,
	}, nil
}

// sessionStore returns a new session store with a random auth and encryption key.
// This means that cookies using the store will be reset if gotosocial is restarted!
func sessionStore() (memstore.Store, error) {
	auth := make([]byte, 32)
	crypt := make([]byte, 32)

	if _, err := rand.Read(auth); err != nil {
		return nil, err
	}
	if _, err := rand.Read(crypt); err != nil {
		return nil, err
	}

	return memstore.NewStore(auth, crypt), nil
}
