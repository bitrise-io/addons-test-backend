package actions

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bitrise-io/api-utils/httpresponse"

	"github.com/bitrise-io/addons-test-backend/bitrise"
	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// LoginPostHandler ...
func LoginPostHandler(appEnv *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	timestamp := r.FormValue("timestamp")
	token := r.FormValue("token")
	appSlug := r.FormValue("app_slug")
	requestParams := appEnv.RequestParams.Get(r)
	buildSlug := requestParams["build_slug"]
	appTitle := requestParams["app_title"]

	appEnv.Logger.Info("Login form data",
		zap.String("timestamp", timestamp),
		zap.String("token", token),
		zap.String("app_slug", appSlug),
		zap.String("build_slug", buildSlug),
	)

	appSlugStored, ok := appEnv.Session.Get("app_slug").(string)
	if ok {
		if appSlug == appSlugStored {
			if buildSlug == "" {
				var err error
				buildSlug, err = fetchBuildSlug(appEnv, appSlug)
				if err != nil {
					return errors.WithMessage(err, "Failed to fetch latest build slug for app")
				}
			}
			http.Redirect(w, r, fmt.Sprintf("/builds/%s", buildSlug), http.StatusMovedPermanently)
			return nil
		}
	}

	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.WithMessage(err, "Failed to parse timestamp int")
	}
	tm := time.Unix(i, 0)

	if time.Now().After(tm.Add(5 * time.Minute)) {
		return httpresponse.RespondWithForbidden(w)
	}

	hashPrefix := "sha256-"
	var hash hash.Hash
	if strings.HasPrefix(token, hashPrefix) {
		token = strings.TrimPrefix(token, hashPrefix)
		hash = sha256.New()
	} else {
		hash = sha1.New()
	}

	_, err = hash.Write([]byte(fmt.Sprintf("%s:%s:%s", appSlug, appEnv.SSOToken, timestamp)))
	if err != nil {
		return errors.WithMessage(err, "Failed to write into sha1 buffer")
	}
	refToken := fmt.Sprintf("%x", hash.Sum(nil))

	if token != refToken {
		appEnv.Logger.Error("Token mismatch")
		appEnv.Session.Clear()
		return httpresponse.RespondWithForbidden(w)
	}

	appEnv.Session.Set("app_slug", appSlug)
	appEnv.Session.Set("app_title", appTitle)

	err = appEnv.Session.Save()
	if err != nil {
		return errors.WithMessage(err, "Failed to save session")
	}

	if buildSlug == "" {
		var err error
		buildSlug, err = fetchBuildSlug(appEnv, appSlug)
		if err != nil {
			return errors.WithMessage(err, "Failed to fetch latest build slug for app")
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/builds/%s", buildSlug), http.StatusMovedPermanently)
	return nil
}

func fetchBuildSlug(appEnv *env.AppEnv, appSlug string) (string, error) {
	app, err := appEnv.AppService.Find(&models.App{AppSlug: appSlug})
	if err != nil {
		return "", errors.WithMessage(err, "SQL Error")
	}
	bc := bitrise.NewClient(app.APIToken)
	build, err := bc.GetLatestBuildOfApp(appSlug)
	if err != nil {
		return "", err
	}
	return build.Slug, nil
}
