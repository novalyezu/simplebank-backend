package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/novalyezu/simplebank-backend/db/sqlc"
	"github.com/novalyezu/simplebank-backend/util"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=3"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

func (server *Server) createUser(c *gin.Context) {
	var body createUserRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       body.Username,
		HashedPassword: hashedPassword,
		FullName:       body.FullName,
		Email:          body.Email,
	}

	user, err := server.store.CreateUser(c, arg)
	if err != nil {
		if pgError, ok := err.(*pq.Error); ok {
			switch pgError.Constraint {
			case "users_pkey":
				c.JSON(http.StatusForbidden, errorResponse(fmt.Errorf("username already exists")))
				return
			case "users_email_key":
				c.JSON(http.StatusForbidden, errorResponse(fmt.Errorf("email already exists")))
				return
			}
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, user)
}
