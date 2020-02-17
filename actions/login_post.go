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

	"github.com/bitrise-io/addons-firebase-testlab/analyticsutils"
	"github.com/bitrise-io/addons-firebase-testlab/bitrise"
	"github.com/bitrise-io/addons-firebase-testlab/configs"
	"github.com/bitrise-io/addons-firebase-testlab/database"
	"github.com/bitrise-io/addons-test-backend/env"
	"github.com/bitrise-io/addons-test-backend/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// LoginPostHandler ...
func LoginPostHandler(env *env.AppEnv, w http.ResponseWriter, r *http.Request) error {
	timestamp := r.FormValue("timestamp")
	token := r.FormValue("token")
	appSlug := r.FormValue("app_slug")
	requestParams := env.RequestParams.Get(r)
	buildSlug := requestParams["build_slug"]
	appTitle := requestParams["app_title"]

	env.Logger.Info("Login form data",
		zap.String("timestamp", timestamp),
		zap.String("token", token),
		zap.String("app_slug", appSlug),
		zap.String("build_slug", buildSlug),
	)

	analyticsutils.SendAddonEvent(analyticsutils.EventAddonSSOLogin, appSlug, "", "")

	session, err := env.SessionStore.Get(r, env.SessionName)
	if err != nil {
		return errors.Wrap(err, "Failed to get session store")
	}
	appSlugStored := session.Values["app_slug"]
	if appSlugStored != "" && appSlug == appSlugStored {
		if buildSlug == "" {
			var err error
			buildSlug, err = fetchBuildSlug(appSlug)
			if err != nil {
				return errors.Wrap(err, "Failed to fetch latest build slug for app")
			}
		}
		http.Redirect(w, r, fmt.Sprintf("/builds/%s", buildSlug), http.StatusMovedPermanently)
		return nil
	}

	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.Wrap(err, "Failed to parse timestamp int")
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

	_, err = hash.Write([]byte(fmt.Sprintf("%s:%s:%s", appSlug, configs.GetAddonSSOToken(), timestamp)))
	if err != nil {
		return errors.Wrap(err, "Failed to write into sha1 buffer")
	}
	refToken := fmt.Sprintf("%x", hash.Sum(nil))

	if token != refToken {
		env.Logger.Error("Token mismatch")
		env.Session.Clear()
		return httpresponse.RespondWithForbidden(w)
	}

	c.Session().Set("app_slug", appSlug)
	c.Session().Set("app_title", appTitle)

	err = c.Session().Save()
	if err != nil {
		logger.Error("Failed to save session", zap.Any("error", errors.WithStack(err)))
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
	}

	if buildSlug == "" {
		var err error
		buildSlug, err = fetchBuildSlug(appSlug)
		if err != nil {
			logger.Error("Failed to fetch latest build slug for app", zap.Error(err))
			return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Internal error"}))
		}
	}

	return c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/builds/%s", buildSlug))
}

func fetchBuildSlug(appSlug string) (string, error) {
	app, err := database.GetApp(&models.App{AppSlug: appSlug})
	if err != nil {
		return "", errors.WithStack(err)
	}
	bc := bitrise.NewClient(app.APIToken)
	build, err := bc.GetLatestBuildOfApp(appSlug)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return build.Slug, nil
}
