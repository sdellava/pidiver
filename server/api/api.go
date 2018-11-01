package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type Request struct {
	Command      string
	Hashes       []string
	Uris         []string
	Addresses    []string
	Bundles      []string
	Tags         []string
	Approvees    []string
	Transactions []string
	Trytes       []string
	Timestamp    int
	// for attachToTangle
	TrunkTransaction   string
	BranchTransaction  string
	MinWeightMagnitude int
}

var api *gin.Engine
var srv *http.Server
var config *viper.Viper
var limitAccess []string
var authEnabled = false
var dummyHash = strings.Repeat("9", 81)
var apiCalls = make(map[string]func(request Request, c *gin.Context, t time.Time))
var startModules []func(apiConfig *viper.Viper)

// TODO: Add attach/interrupt attaching api
// TODO: limit requests, lists, etc.

func Start(apiConfig *viper.Viper) {
	config = apiConfig
	if !config.GetBool("api.debug") {
		gin.SetMode(gin.ReleaseMode)
	}
	limitAccess = config.GetStringSlice("api.limitRemoteAccess")
	Log.Debug("Limited remote access to:", limitAccess)

	// pass config to modules if they need it
	for _, f := range startModules {
		f(apiConfig)
	}

	api = gin.Default()

	username := config.GetString("api.auth.username")
	password := config.GetString("api.auth.password")
	if len(username) > 0 && len(password) > 0 {
		api.Use(gin.BasicAuth(gin.Accounts{username: password}))
	}

	api.POST("/", func(c *gin.Context) {
		t := time.Now()

		var request Request
		err := c.ShouldBindJSON(&request)
		if err == nil {

			if triesToAccessLimited(request.Command, c) {
				Log.Warningf("Denying limited command request %v from remote %v",
					request.Command, c.Request.RemoteAddr)
				ReplyError("Limited remote command access", c)
				return
			}

			caseInsensitiveCommand := strings.ToLower(request.Command)
			apiCall, apiCallExists := apiCalls[caseInsensitiveCommand]
			if apiCallExists {
				apiCall(request, c, t)
			} else {
				Log.Error("Unknown command", request.Command)
				ReplyError("No known command provided", c)
			}

		} else {
			Log.Error("ERROR request", err)
			ReplyError("Wrongly formed JSON", c)
		}
	})

	srv = &http.Server{
		Addr:    config.GetString("api.host") + ":" + config.GetString("api.port"),
		Handler: api,
	}
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Log.Fatal("API Server Error", err)
		}
	}()
}

func End() {
	if srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			Log.Fatal("API Server Shutdown Error:", err)
		}
		Log.Info("API Server exiting...")
	}
}

func ReplyError(message string, c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": message,
	})
}

func getDuration(t time.Time) float64 {
	return time.Now().Sub(t).Seconds()
}

func triesToAccessLimited(command string, c *gin.Context) bool {
	if c.Request.RemoteAddr[:9] == "127.0.0.1" {
		return false
	}
	for _, l := range limitAccess {
		if l == command {
			return true
		}
	}
	return false
}

func addAPICall(apiCall string, implementation func(request Request, c *gin.Context, t time.Time)) {
	caseInsensitiveApiCall := strings.ToLower(apiCall)
	apiCalls[caseInsensitiveApiCall] = implementation
}

func addStartModule(implementation func(apiConfig *viper.Viper)) {
	startModules = append(startModules, implementation)
}
