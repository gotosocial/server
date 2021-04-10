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

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config pulls together all the configuration needed to run gotosocial
type Config struct {
	LogLevel        string          `yaml:"logLevel"`
	ApplicationName string          `yaml:"applicationName"`
	Host            string          `yaml:"host"`
	Protocol        string          `yaml:"protocol"`
	DBConfig        *DBConfig       `yaml:"db"`
	TemplateConfig  *TemplateConfig `yaml:"template"`
	AccountsConfig  *AccountsConfig `yaml:"accounts"`
	MediaConfig     *MediaConfig    `yaml:"media"`
	StorageConfig   *StorageConfig  `yaml:"storage"`
	StatusesConfig  *StatusesConfig `yaml:"statuses"`
}

// FromFile returns a new config from a file, or an error if something goes amiss.
func FromFile(path string) (*Config, error) {
	if path != "" {
		c, err := loadFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("error creating config: %s", err)
		}
		return c, nil
	}
	return Empty(), nil
}

// Empty just returns a new empty config
func Empty() *Config {
	return &Config{
		DBConfig:       &DBConfig{},
		TemplateConfig: &TemplateConfig{},
		AccountsConfig: &AccountsConfig{},
		MediaConfig:    &MediaConfig{},
		StorageConfig:  &StorageConfig{},
		StatusesConfig: &StatusesConfig{},
	}
}



// loadFromFile takes a path to a yaml file and attempts to load a Config object from it
func loadFromFile(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file at path %s: %s", path, err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(bytes, config); err != nil {
		return nil, fmt.Errorf("could not unmarshal file at path %s: %s", path, err)
	}

	return config, nil
}

// ParseCLIFlags sets flags on the config using the provided Flags object
func (c *Config) ParseCLIFlags(f KeyedFlags) {
	fn := GetFlagNames()

	// For all of these flags, we only want to set them on the config if:
	//
	// a) They haven't been set at all in the config file we already parsed,
	// 	  and so we take the default from the flags object.
	//
	// b) They may have been set in the config, but they've *also* been set explicitly
	//    as a command-line argument or an env variable, which takes priority.

	// general flags
	if c.LogLevel == "" || f.IsSet(fn.LogLevel) {
		c.LogLevel = f.String(fn.LogLevel)
	}

	if c.ApplicationName == "" || f.IsSet(fn.ApplicationName) {
		c.ApplicationName = f.String(fn.ApplicationName)
	}

	if c.Host == "" || f.IsSet(fn.Host) {
		c.Host = f.String(fn.Host)
	}

	if c.Protocol == "" || f.IsSet(fn.Protocol) {
		c.Protocol = f.String(fn.Protocol)
	}

	// db flags
	if c.DBConfig.Type == "" || f.IsSet(fn.DbType) {
		c.DBConfig.Type = f.String(fn.DbType)
	}

	if c.DBConfig.Address == "" || f.IsSet(fn.DbAddress) {
		c.DBConfig.Address = f.String(fn.DbAddress)
	}

	if c.DBConfig.Port == 0 || f.IsSet(fn.DbPort) {
		c.DBConfig.Port = f.Int(fn.DbPort)
	}

	if c.DBConfig.User == "" || f.IsSet(fn.DbUser) {
		c.DBConfig.User = f.String(fn.DbUser)
	}

	if c.DBConfig.Password == "" || f.IsSet(fn.DbPassword) {
		c.DBConfig.Password = f.String(fn.DbPassword)
	}

	if c.DBConfig.Database == "" || f.IsSet(fn.DbDatabase) {
		c.DBConfig.Database = f.String(fn.DbDatabase)
	}

	// template flags
	if c.TemplateConfig.BaseDir == "" || f.IsSet(fn.TemplateBaseDir) {
		c.TemplateConfig.BaseDir = f.String(fn.TemplateBaseDir)
	}

	// accounts flags
	if f.IsSet(fn.AccountsOpenRegistration) {
		c.AccountsConfig.OpenRegistration = f.Bool(fn.AccountsOpenRegistration)
	}

	if f.IsSet(fn.AccountsApprovalRequired) {
		c.AccountsConfig.RequireApproval = f.Bool(fn.AccountsApprovalRequired)
	}

	// media flags
	if c.MediaConfig.MaxImageSize == 0 || f.IsSet(fn.MediaMaxImageSize) {
		c.MediaConfig.MaxImageSize = f.Int(fn.MediaMaxImageSize)
	}

	if c.MediaConfig.MaxVideoSize == 0 || f.IsSet(fn.MediaMaxVideoSize) {
		c.MediaConfig.MaxVideoSize = f.Int(fn.MediaMaxVideoSize)
	}

	if c.MediaConfig.MinDescriptionChars == 0 || f.IsSet(fn.MediaMinDescriptionChars) {
		c.MediaConfig.MinDescriptionChars = f.Int(fn.MediaMinDescriptionChars)
	}

	if c.MediaConfig.MaxDescriptionChars == 0 || f.IsSet(fn.MediaMaxDescriptionChars) {
		c.MediaConfig.MaxDescriptionChars = f.Int(fn.MediaMaxDescriptionChars)
	}

	// storage flags
	if c.StorageConfig.Backend == "" || f.IsSet(fn.StorageBackend) {
		c.StorageConfig.Backend = f.String(fn.StorageBackend)
	}

	if c.StorageConfig.BasePath == "" || f.IsSet(fn.StorageBasePath) {
		c.StorageConfig.BasePath = f.String(fn.StorageBasePath)
	}

	if c.StorageConfig.ServeProtocol == "" || f.IsSet(fn.StorageServeProtocol) {
		c.StorageConfig.ServeProtocol = f.String(fn.StorageServeProtocol)
	}

	if c.StorageConfig.ServeHost == "" || f.IsSet(fn.StorageServeHost) {
		c.StorageConfig.ServeHost = f.String(fn.StorageServeHost)
	}

	if c.StorageConfig.ServeBasePath == "" || f.IsSet(fn.StorageServeBasePath) {
		c.StorageConfig.ServeBasePath = f.String(fn.StorageServeBasePath)
	}

	// statuses flags
	if c.StatusesConfig.MaxChars == 0 || f.IsSet(fn.StatusesMaxChars) {
		c.StatusesConfig.MaxChars = f.Int(fn.StatusesMaxChars)
	}
	if c.StatusesConfig.CWMaxChars == 0 || f.IsSet(fn.StatusesCWMaxChars) {
		c.StatusesConfig.CWMaxChars = f.Int(fn.StatusesCWMaxChars)
	}
	if c.StatusesConfig.PollMaxOptions == 0 || f.IsSet(fn.StatusesPollMaxOptions) {
		c.StatusesConfig.PollMaxOptions = f.Int(fn.StatusesPollMaxOptions)
	}
	if c.StatusesConfig.PollOptionMaxChars == 0 || f.IsSet(fn.StatusesPollOptionMaxChars) {
		c.StatusesConfig.PollOptionMaxChars = f.Int(fn.StatusesPollOptionMaxChars)
	}
	if c.StatusesConfig.MaxMediaFiles == 0 || f.IsSet(fn.StatusesMaxMediaFiles) {
		c.StatusesConfig.MaxMediaFiles = f.Int(fn.StatusesMaxMediaFiles)
	}
}

// KeyedFlags is a wrapper for any type that can store keyed flags and give them back.
// HINT: This works with a urfave cli context struct ;)
type KeyedFlags interface {
	Bool(k string) bool
	String(k string) string
	Int(k string) int
	IsSet(k string) bool
}

// Flags is used for storing the names of the various flags used for
// initializing and storing urfavecli flag variables.
type Flags struct {
	LogLevel        string
	ApplicationName string
	ConfigPath      string
	Host            string
	Protocol        string

	DbType     string
	DbAddress  string
	DbPort     string
	DbUser     string
	DbPassword string
	DbDatabase string

	TemplateBaseDir string

	AccountsOpenRegistration string
	AccountsApprovalRequired string
	AccountsReasonRequired   string

	MediaMaxImageSize        string
	MediaMaxVideoSize        string
	MediaMinDescriptionChars string
	MediaMaxDescriptionChars string

	StorageBackend       string
	StorageBasePath      string
	StorageServeProtocol string
	StorageServeHost     string
	StorageServeBasePath string

	StatusesMaxChars           string
	StatusesCWMaxChars         string
	StatusesPollMaxOptions     string
	StatusesPollOptionMaxChars string
	StatusesMaxMediaFiles      string
}

type Defaults struct {
	LogLevel        string
	ApplicationName string
	ConfigPath      string
	Host            string
	Protocol        string

	DbType     string
	DbAddress  string
	DbPort     int
	DbUser     string
	DbPassword string
	DbDatabase string

	TemplateBaseDir string

	AccountsOpenRegistration bool
	AccountsRequireApproval  bool
	AccountsReasonRequired   bool

	MediaMaxImageSize        int
	MediaMaxVideoSize        int
	MediaMinDescriptionChars int
	MediaMaxDescriptionChars int

	StorageBackend       string
	StorageBasePath      string
	StorageServeProtocol string
	StorageServeHost     string
	StorageServeBasePath string

	StatusesMaxChars           int
	StatusesCWMaxChars         int
	StatusesPollMaxOptions     int
	StatusesPollOptionMaxChars int
	StatusesMaxMediaFiles      int
}

// GetFlagNames returns a struct containing the names of the various flags used for
// initializing and storing urfavecli flag variables.
func GetFlagNames() Flags {
	return Flags{
		LogLevel:        "log-level",
		ApplicationName: "application-name",
		ConfigPath:      "config-path",
		Host:            "host",
		Protocol:        "protocol",

		DbType:     "db-type",
		DbAddress:  "db-address",
		DbPort:     "db-port",
		DbUser:     "db-user",
		DbPassword: "db-password",
		DbDatabase: "db-database",

		TemplateBaseDir: "template-basedir",

		AccountsOpenRegistration: "accounts-open-registration",
		AccountsApprovalRequired: "accounts-approval-required",
		AccountsReasonRequired:   "accounts-reason-required",

		MediaMaxImageSize:        "media-max-image-size",
		MediaMaxVideoSize:        "media-max-video-size",
		MediaMinDescriptionChars: "media-min-description-chars",
		MediaMaxDescriptionChars: "media-max-description-chars",

		StorageBackend:       "storage-backend",
		StorageBasePath:      "storage-base-path",
		StorageServeProtocol: "storage-serve-protocol",
		StorageServeHost:     "storage-serve-host",
		StorageServeBasePath: "storage-serve-base-path",

		StatusesMaxChars:           "statuses-max-chars",
		StatusesCWMaxChars:         "statuses-cw-max-chars",
		StatusesPollMaxOptions:     "statuses-poll-max-options",
		StatusesPollOptionMaxChars: "statuses-poll-option-max-chars",
		StatusesMaxMediaFiles:      "statuses-max-media-files",
	}
}

// GetEnvNames returns a struct containing the names of the environment variable keys used for
// initializing and storing urfavecli flag variables.
func GetEnvNames() Flags {
	return Flags{
		LogLevel:        "GTS_LOG_LEVEL",
		ApplicationName: "GTS_APPLICATION_NAME",
		ConfigPath:      "GTS_CONFIG_PATH",
		Host:            "GTS_HOST",
		Protocol:        "GTS_PROTOCOL",

		DbType:     "GTS_DB_TYPE",
		DbAddress:  "GTS_DB_ADDRESS",
		DbPort:     "GTS_DB_PORT",
		DbUser:     "GTS_DB_USER",
		DbPassword: "GTS_DB_PASSWORD",
		DbDatabase: "GTS_DB_DATABASE",

		TemplateBaseDir: "GTS_TEMPLATE_BASEDIR",

		AccountsOpenRegistration: "GTS_ACCOUNTS_OPEN_REGISTRATION",
		AccountsApprovalRequired: "GTS_ACCOUNTS_APPROVAL_REQUIRED",
		AccountsReasonRequired:   "GTS_ACCOUNTS_REASON_REQUIRED",

		MediaMaxImageSize:        "GTS_MEDIA_MAX_IMAGE_SIZE",
		MediaMaxVideoSize:        "GTS_MEDIA_MAX_VIDEO_SIZE",
		MediaMinDescriptionChars: "GTS_MEDIA_MIN_DESCRIPTION_CHARS",
		MediaMaxDescriptionChars: "GTS_MEDIA_MAX_DESCRIPTION_CHARS",

		StorageBackend:       "GTS_STORAGE_BACKEND",
		StorageBasePath:      "GTS_STORAGE_BASE_PATH",
		StorageServeProtocol: "GTS_STORAGE_SERVE_PROTOCOL",
		StorageServeHost:     "GTS_STORAGE_SERVE_HOST",
		StorageServeBasePath: "GTS_STORAGE_SERVE_BASE_PATH",

		StatusesMaxChars:           "GTS_STATUSES_MAX_CHARS",
		StatusesCWMaxChars:         "GTS_STATUSES_CW_MAX_CHARS",
		StatusesPollMaxOptions:     "GTS_STATUSES_POLL_MAX_OPTIONS",
		StatusesPollOptionMaxChars: "GTS_STATUSES_POLL_OPTION_MAX_CHARS",
		StatusesMaxMediaFiles:      "GTS_STATUSES_MAX_MEDIA_FILES",
	}
}
