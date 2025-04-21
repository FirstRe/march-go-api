package authService

import (
	"bytes"
	utils "core"
	jwts "core/app/common/jwt"
	"core/app/helper"
	"encoding/json"
	"fmt"
	"log"
	gormDb "march-auth/cmd/app/common/gorm"
	config "march-auth/cmd/app/constant"
	"march-auth/cmd/app/graph/model"
	"march-auth/cmd/app/graph/types"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

const ClassNameAuth string = "AuthService"

type dataOAuth struct {
	Access_token  string `json:"access_token"`
	Expires_in    int    `json:"expires_in"`
	Refresh_token string `json:"refresh_token"`
	Scope         string `json:"scope"`
	Token_type    string `json:"token_type"`
	Id_token      string `json:"id_token"`
}

func SignInBypass() (*types.Token, error) {
	logctx := helper.LogContext(ClassNameAuth, "SignInBypass")

	findFirst := &model.User{}
	gormDb.Repos.Model(&model.User{}).
		Preload(clause.Associations).
		Preload("Group.GroupFunctions").
		Preload("Group.GroupFunctions.Function").
		Preload("Group.GroupTasks").
		Preload("Group.GroupTasks.Task").
		Preload("Group.Shop").
		Where("email = ?", "firstzaxshot95@gmail.com").Find(findFirst)
		logctx.Loggers("findFirst", findFirst)
	return genToken(findFirst)
}

func SignInOAuth(code string) (*types.Token, error) {
	logctx := helper.LogContext(ClassNameAuth, "OAuthURL")
	logctx.Logger(code, "code")
	if code == "" {
		return nil, fmt.Errorf("no Code")
	}

	data := url.Values{}
	data.Set("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("GOOGLE_SECRET"))
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", os.Getenv("REDIRECT_URL"))

	logctx.Logger(data, "data")

	req, err := http.NewRequest("POST", config.ConfigOAuth.TokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("Error creating request:%v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Error sending request:%v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Err api:%v", resp.StatusCode)
	}

	defer resp.Body.Close()
	log.Printf("resp.Body: %+v", resp.StatusCode)

	responseData := &dataOAuth{}

	derr := json.NewDecoder(resp.Body).Decode(responseData)

	if derr != nil {
		return nil, fmt.Errorf("Decode error:%v", derr)
	}

	logctx.Logger(responseData, "responseData")

	token, err := jwt.Parse(responseData.Id_token, nil)

	if token == nil {
		return nil, err
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	logctx.Logger(claims, "claims")

	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	picture, _ := claims["picture"].(string)

	logctx.Logger(email, "emailO")

	revokeToken(responseData.Access_token, responseData.Refresh_token)

	findFirst := &model.User{}
	gormDb.Repos.Model(&model.User{}).
		Preload(clause.Associations).
		Preload("Group.GroupFunctions").
		Preload("Group.GroupFunctions.Function").
		Preload("Group.GroupTasks").
		Preload("Group.GroupTasks.Task").
		Preload("Group.Shop").
		Where("email = ?", email).Find(findFirst)

	logctx.Logger(findFirst, "findFirst")

	if findFirst.ID == "" || findFirst.Deleted {
		return nil, fmt.Errorf("Unauthorized No Access")
	}

	gormDb.Repos.Model(&model.User{}).
		Where("id = ?", findFirst.ID).
		Updates(&model.User{
			Username:     name,
			IsRegistered: utils.BoolAddr(true),
			Picture:      &picture,
		})

	return genToken(findFirst)
}

func VerifyAccessToken(token string) (*types.VerifyAccessTokenResponse, error) {
	logctx := helper.LogContext(ClassNameAuth, "VerifyAccessToken")

	pass := true
	validate, err := jwts.Verify(token)

	if err != nil || !validate.Valid {
		pass = false
	}

	response := types.VerifyAccessTokenResponse{
		Success: utils.BoolAddr(pass),
	}

	logctx.Logger(pass, "isPass")
	return &response, nil
}

func SignOut(id string) (*types.SignOutResponse, error) {
	logctx := helper.LogContext(ClassNameAuth, "SignOut")

	user := &model.User{}
	if err := gormDb.Repos.
		Where("id = ?", id).Model(user).
		Updates(map[string]interface{}{"refresh_token": nil, "device_id": nil}).
		Error; err != nil {
		logctx.Logger(err.Error(), "[error-api] SignOut")
		return nil, fmt.Errorf("Internal Server Error: %v", err)
	}
	logctx.Logger(user, "user")
	return &types.SignOutResponse{
		ID: id,
	}, nil
}

func TokenExpire(refreshToken string) (*types.Token, error) {
	logctx := helper.LogContext(ClassNameAuth, "TokenExpire")
	logctx.Logger(refreshToken, "refreshToken")

	verify, err := jwts.Verify(refreshToken, true)

	if err != nil || !verify.Valid {
		logctx.Logger(nil, "[log-err] InValid")
		return nil, fmt.Errorf("Unauthorized")
	}

	decode, err := jwts.DecodeRefresh(refreshToken)

	if err != nil {
		logctx.Logger(nil, "[log-err] decode")
		return nil, fmt.Errorf("Unauthorized")
	}

	user := &model.User{}
	gormDb.Repos.Model(&model.User{}).
		Preload(clause.Associations).
		Preload("Group.GroupFunctions").
		Preload("Group.GroupFunctions.Function").
		Preload("Group.GroupTasks").
		Preload("Group.GroupTasks.Task").
		Preload("Group.Shop").
		Where("id = ?", decode.ID).Find(user)

	if user.RefreshToken != nil && refreshToken != *user.RefreshToken || *user.DeviceID != decode.DeviceId {
		gormDb.Repos.Model(&model.User{}).Where("id = ?", user.ID).Update("device_id", nil)
		return nil, fmt.Errorf("Unauthorized")
	}

	token := genAccessToken(decode.DeviceId, user)

	return &types.Token{
		AccessToken: token,
	}, nil
}

func revokeToken(accessToken, refreshToken string) {
	logctx := helper.LogContext(ClassNameAuth, "revokeToken")
	tokens := []string{accessToken, refreshToken}
	for _, token := range tokens {
		go func(t string) {
			data := []byte(fmt.Sprintf("token=%s", t))
			req, err := http.NewRequest("POST", config.ConfigOAuth.RevokeTokenURL, bytes.NewBuffer(data))
			if err != nil {
				return
			}

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return
			}
		}(token)
	}
	logctx.Logger(ClassNameAuth, "revokeToken")
}

func genToken(user *model.User) (*types.Token, error) {
	logctx := helper.LogContext(ClassNameAuth, "genToken")
	deviceId := uuid.New().String()

	logctx.Logger(deviceId, "deviceId")
	token := genAccessToken(deviceId, user)

	jwtRefData := &jwts.JwtCustomClaimRef{
		ID:       user.ID,
		DeviceId: deviceId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7 * 4 * 99).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	refresh_token, _ := jwts.JwtGenerateRef(jwtRefData)
	//update deviceId refreshToken

	gormDb.Repos.Where("id = ?", user.ID).Updates(&model.User{RefreshToken: &refresh_token, DeviceID: &deviceId})

	reponse := types.Token{
		AccessToken:  token,
		RefreshToken: &refresh_token,
		Username:     &user.Username,
		UserID:       &user.ID,
	}
	return &reponse, nil
}

func genAccessToken(deviceId string, user *model.User) string {
	logctx := helper.LogContext(ClassNameAuth, "genAccessToken")
	page := addPage(user.Group.GroupFunctions)
	Role := strings.ToUpper(strings.Split(user.Group.Name, "|")[0])
	var tasks []string

	for _, item := range user.Group.GroupTasks {
		tasks = append(tasks, item.Task.Name)
	}

	var functions []string

	for _, item := range user.Group.GroupFunctions {
		functions = append(functions, item.Function.Name)
	}

	jwtData := &jwts.JwtCustomClaim{
		Role:     Role,
		DeviceId: deviceId,
		UserId:   user.ID,
		ShopsID:  user.ShopsID,
		ShopName: user.Shop.Name,
		UserName: user.Username,
		Picture:  *user.Picture,
		Info: jwts.Info{
			Page:      page,
			Tasks:     tasks,
			Functions: functions,
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	logctx.Logger(jwtData, "jwtData")

	token, _ := jwts.JwtGenerateAcc(jwtData)
	logctx.Logger(token, "token")
	return token
}

func addPage(groupFunctions []model.GroupFunction) map[string][]string {
	result := make(map[string][]string)

	for _, item := range groupFunctions {
		name := item.Function.Name
		var values []string
		if item.Create {
			values = append(values, "CREATE")
		}
		if item.View {
			values = append(values, "VIEW")
		}
		if item.Update {
			values = append(values, "UPDATE")
		}
		result[name] = values
	}

	return result
}
