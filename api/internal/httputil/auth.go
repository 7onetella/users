package httputil

import "github.com/7onetella/users/api/internal/model"

func (rh RequestHandler) NewAuthEvent(userID, event string) model.AuthEvent {
	c := rh.Context

	return model.NewAuthEvent(userID, event, c.ClientIP(), c.ClientIP(), c.Request.UserAgent())
}
