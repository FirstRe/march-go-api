package jwt

import (
	"context"
	"core/app/helper"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type authString string

type Info struct {
	Tasks []string `json:"tasks"`
}

type JwtCustomClaim struct {
	ShopsID  string `json:"shopsId"`
	Role     string `json:"role"`
	UserId   string `json:"userId"`
	DeviceId string `json:"deviceId"`
	ShopName string `json:"shopName"`
	UserName string `json:"userName"`
	Info     Info   `json:"info"`
	jwt.StandardClaims
}

var jwtSecret = []byte(getJwtSecret())

func getJwtSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "secret_MAKMAK"
	}
	return secret
}

func JwtGenerate(shopsID string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &JwtCustomClaim{
		ShopsID: shopsID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	})

	token, err := t.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}

func JwtValidate(ctx context.Context, token string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(token, &JwtCustomClaim{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("there's a problem with the signing method")
		}
		return jwtSecret, nil
	})
}

func VerifyJWT(tokens string) (*JwtCustomClaim, error) {
	logctx := helper.LogContext("JWT", "VerifyJWT")

	token, err := jwt.Parse(tokens, nil)

	if token == nil {
		return nil, err
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	logctx.Logger(claims, "claims")

	shopsId, _ := claims["shopsId"].(string)
	role, _ := claims["role"].(string)
	userId, _ := claims["userId"].(string)
	deviceId, _ := claims["deviceId"].(string)
	shopName, _ := claims["shopName"].(string)
	userName, _ := claims["userName"].(string)
	infoMap, _ := claims["info"].(map[string]interface{})

	infoJson, err := json.Marshal(infoMap)
	if err != nil {
		return nil, err
	}

	var info Info
	err = json.Unmarshal(infoJson, &info)

	if err != nil {
		return nil, err
	}

	userInfo := JwtCustomClaim{
		ShopsID:  shopsId,
		Role:     role,
		UserId:   userId,
		Info:     info,
		DeviceId: deviceId,
		ShopName: shopName,
		UserName: userName,
	}

	return &userInfo, nil
}
