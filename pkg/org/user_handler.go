package org

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httperror"
	"github.com/uptrace/uptrace/pkg/httputil"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	*bunapp.App
}

func NewUserHandler(app *bunapp.App) *UserHandler {
	return &UserHandler{
		App: app,
	}
}

func (h *UserHandler) Current(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := UserFromContext(ctx)

	projects, err := SelectProjects(ctx, h.App)
	if err != nil {
		return err
	}

	return httputil.JSON(w, bunrouter.H{
		"user":     user,
		"projects": projects,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
		return err
	}

	user, err := SelectUserByUsername(ctx, h.App, in.Username)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return httperror.BadRequest("credentials", "user with such credentials not found")
	}

	token, err := encodeUserToken(h.Config().SecretKey, user.Username, tokenTTL)
	if err != nil {
		return err
	}

	cookie := bunapp.NewCookie(h.App, req)
	cookie.Name = tokenCookieName
	cookie.Value = token
	cookie.MaxAge = int(tokenTTL.Seconds())
	http.SetCookie(w, cookie)

	return nil
}

func (h *UserHandler) Logout(w http.ResponseWriter, req bunrouter.Request) error {
	cookie := bunapp.NewCookie(h.App, req)
	cookie.Name = tokenCookieName
	cookie.Expires = time.Now().Add(-time.Hour)
	http.SetCookie(w, cookie)

	return nil
}

func (h *UserHandler) Oauth(w http.ResponseWriter, req bunrouter.Request) error {
	conf := h.App.Config()
	query := req.URL.Query()
	redirect := query["redirect"]
	ticket := query["ticket"]

	service := fmt.Sprintf("%s/api/v1/users/oauth?redirect=%s", conf.Site.Addr, redirect[0])
	user, err := h.checkToken(req.Context(), service, ticket[0])
	if err != nil {
		return err
	}

	token, err := encodeUserToken(h.Config().SecretKey, user.Username, tokenTTL)
	if err != nil {
		return err
	}

	cookie := bunapp.NewCookie(h.App, req)
	cookie.Name = tokenCookieName
	cookie.Value = token
	cookie.MaxAge = int(tokenTTL.Seconds())
	http.SetCookie(w, cookie)
	http.Redirect(w, req.Request, redirect[0], http.StatusFound)
	return nil
}

func (h *UserHandler) checkToken(ctx context.Context, service, token string) (*User, error) {
	conf := h.App.Config()
	tokenUrl := fmt.Sprintf("%s%s", conf.Auth.Oauth.Host, conf.Auth.Oauth.TokenPath)
	param := url.Values{}
	param.Add("service", service)
	param.Add("ticket", token)
	param.Add("format", "JSON")

	resp, err := http.Get(tokenUrl + "?" + param.Encode())
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]interface{})
	if err := json.Unmarshal(b, &ret); err != nil {
		return nil, err
	}

	if response, ok := ret["serviceResponse"].(map[string]interface{}); ok {
		if succ, ok := response["authenticationSuccess"].(map[string]interface{}); ok {
			username := succ["attributes"].(map[string]interface{})["name"].(string)
			email := succ["attributes"].(map[string]interface{})["email"].(string)
			user, err := SelectUserByUsername(ctx, h.App, username)
			if err != nil {
				if err.Error() == "sql: no rows in result set" {
					dest := &User{
						Username: username,
						Email:    email,
					}
					if err := dest.SetPassword(fmt.Sprintf("%s_uptrace", email)); err != nil {
						return nil, err
					}
					if err := dest.Init(); err != nil {
						return nil, err
					}

					if _, err := h.PG.NewInsert().
						Model(dest).
						On("CONFLICT (username) DO UPDATE").
						Set("password = EXCLUDED.password").
						Set("avatar = EXCLUDED.avatar").
						Exec(ctx); err != nil {
						return nil, err
					}
					return dest, nil
				} else {
					return nil, err
				}
			}
			return user, nil
		}
	}
	return nil, errors.New("invalid token")
}
