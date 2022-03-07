package routes

import (
	"context"
	"fmt"
	"github.com/shaj13/libcache"
	"net/http"
	"time"

	"github.com/shaj13/go-guardian/v2/auth"
	"github.com/shaj13/go-guardian/v2/auth/strategies/basic"
	_ "github.com/shaj13/libcache/fifo"
)

var strategy auth.Strategy
var cacheObj libcache.Cache

func SetupGoGuardian() {
	cacheObj = libcache.FIFO.New(0)
	cacheObj.SetTTL(time.Hour * 24)
	cacheObj.RegisterOnExpired(func(key, _ interface{}) {
		cacheObj.Peek(key)
	})
	strategy = basic.NewCached(validateUser, cacheObj)
}

func validateUser(ctx context.Context, r *http.Request, userName, password string) (auth.Info, error) {
	// here connect to db or any other service to fetch user and validate it.
	if userName == "aerospike" && password == "Aerospike123!" {
		return auth.NewDefaultUser("aerospike", "1", nil, nil), nil
	}

	return nil, fmt.Errorf("[ERROR] Invalid credentials")
}
