package jwt

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"
)

type authString string

type Info struct {
	Functions []string            `json:"functions"`
	Tasks     []string            `json:"tasks"`
	Page      map[string][]string `json:"page"`
}

type JwtCustomClaim struct {
	Role     string `json:"role"`
	Info     Info   `json:"info"`
	DeviceId string `json:"deviceId"`
	UserId   string `json:"userId"`
	ShopsID  string `json:"shopsId"`
	ShopName string `json:"shopName"`
	UserName string `json:"userName"`
	Picture  string `json:"picture"`
	jwt.StandardClaims
}

type JwtCustomClaimRef struct {
	ID       string `json:"id"`
	DeviceId string `json:"deviceId"`
	jwt.StandardClaims
}

var jwtSecret = []byte(getJwtSecret(true))
var jwtSecretRef = []byte(getJwtSecret(false))

func getJwtSecret(acc bool) string {
	if acc == true {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			return "secret_MAKMAK"
		}
		return secret
	} else {
		secret := os.Getenv("JWT_SECRET_REF")
		if secret == "" {
			return "secret_MAKMAKMAK"
		}
		return secret
	}

}

func JwtGenerateAcc(jwts *JwtCustomClaim) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwts)

	token, err := t.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}

func JwtGenerateRef(jwts *JwtCustomClaimRef) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwts)

	token, err := t.SignedString(jwtSecretRef)
	if err != nil {
		return "", err
	}

	return token, nil
}

func Verify(token string, isRefresh ...bool) (*jwt.Token, error) {
	isRefreshNew := false

	if len(isRefresh) > 0 {
		isRefreshNew = isRefresh[0]
	}

	return jwt.ParseWithClaims(token, &JwtCustomClaim{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("there's a problem with the signing method")
		}
		if isRefreshNew {
			return jwtSecretRef, nil
		}
		return jwtSecret, nil
	})
}

func DecodeRefresh(tokens string) (*JwtCustomClaimRef, error) {
	token, err := jwt.Parse(tokens, nil)

	if token == nil {
		return nil, err
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	id, _ := claims["id"].(string)
	deviceId, _ := claims["deviceId"].(string)

	return &JwtCustomClaimRef{
		ID:       id,
		DeviceId: deviceId,
	}, nil

}

func Decode(tokens string) (*JwtCustomClaim, error) {
	// logctx := helper.LogContext("JWT", "VerifyJWT")

	token, err := jwt.Parse(tokens, nil)

	if token == nil {
		return nil, err
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	// logctx.Logger(claims, "claims")

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
