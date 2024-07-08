package token

import (
	"testing"
	"time"

	"github.com/novalyezu/simplebank-backend/util"
	"github.com/stretchr/testify/assert"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	assert.NoError(t, err)

	username := util.RandomString(6)
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, err := maker.CreateToken(username, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	assert.NoError(t, err)
	assert.NotEmpty(t, payload)

	assert.NotZero(t, payload.ID)
	assert.Equal(t, username, payload.Username)
	assert.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	assert.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestPasetoMakerExpiredToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	assert.NoError(t, err)

	username := util.RandomString(6)
	duration := -time.Minute

	token, err := maker.CreateToken(username, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrExpiredToken.Error())
	assert.Nil(t, payload)
}

func TestPasetoMakerInvalidToken(t *testing.T) {
	maker1, err := NewPasetoMaker(util.RandomString(32))
	assert.NoError(t, err)

	maker2, err := NewPasetoMaker(util.RandomString(32))
	assert.NoError(t, err)

	username := util.RandomString(6)
	duration := -time.Minute

	token, err := maker1.CreateToken(username, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	payload, err := maker2.VerifyToken(token)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrInvalidToken.Error())
	assert.Nil(t, payload)
}

func TestPasetoMakerInvalidKeySize(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(16))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "invalid key size")
	assert.Nil(t, maker)
}
