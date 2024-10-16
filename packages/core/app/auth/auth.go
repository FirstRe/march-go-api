package auth

import (
	"bytes"
	"context"
	"core/app/helper"
	"core/app/middlewares"
	"encoding/json"
	"fmt"
	"net/http"

	// "log"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/viper"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type deviceIdPost struct {
	DeviceId string `json:"deviceId"`
}

var uams = NewUAM()

func Auth(ctx context.Context, obj interface{}, next graphql.Resolver, scopes []*string) (interface{}, error) {
	l := helper.LogContext("AuthMiddleware", "Auth")
	userInfo := middlewares.UserInfo(ctx)
	gqlErr := middlewares.GqlErr(ctx)
	accessToken := middlewares.CtxValue(ctx)
	if gqlErr != nil {
		return nil, &gqlerror.Error{
			Message: gqlErr.Message,
		}
	}

	if userInfo.UserInfo.ShopsID == "" {
		return nil, &gqlerror.Error{
			Message: "Access Denied",
		}
	}
	l.Logger(gqlErr, "gqlErr", true)
	l.Logger(scopes, "scopes", true)
	l.Logger(userInfo.UserInfo, "userInfo", true)

	scopes = checkScopeAdmin(scopes)

	if scopes == nil || userInfo.UserInfo.Role == "" {
		if userInfo.UserInfo.Role != "SUPERADMIN" {
			return nil, &gqlerror.Error{
				Message: "Unauthorized Role",
			}
		}
	}
	checkId, err := validateDeviceId(userInfo.UserInfo.DeviceId, accessToken, userInfo.UserInfo.Info.Tasks, scopes)
	l.Logger(checkId, "checkId")

	if !checkId {
		return nil, &gqlerror.Error{
			Message: err,
		}
	}
	return next(ctx)
}

func checkScopeAdmin(scopes []*string) []*string {
	for _, uam := range UAMSTRUCT {
		for _, scope := range scopes {
			if scope != nil && *scope == uam {
				if actions, ok := uams.ScopesMap[uam]; ok {
					newScopes := convertActionsToPointers(actions)
					return newScopes
				}
			}
		}
	}
	return scopes
}

func validateDeviceId(
	deviceIdToken string,
	accessToken string,
	userTask []string,
	scopes []*string) (bool, string) {
	url := viper.GetString("UAM_URL")
	deviceId, err := deviceIdCheckPost(url, accessToken)

	if err != nil {
		return false, "Unauthorized Device"
	}

	isBackOffice := isBackOfficeUser(userTask)

	if deviceIdToken != *deviceId {
		return false, "Unauthorized Device"
	} else {
		if isBackOffice {
			if verifyUser := verifyUserGroups(scopes, userTask); verifyUser {
				return true, ""
			} else {
				return false, "Unauthorized Role"
			}
		} else {
			return false, "Unauthorized not BackOffice"
		}
	}

}

func deviceIdCheckPost(url string, accessToken string) (*string, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(nil))
	if err != nil {
		return nil, fmt.Errorf("Error creating request:", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request:", err)
	}
	defer resp.Body.Close()

	responseData := &deviceIdPost{}

	derr := json.NewDecoder(resp.Body).Decode(responseData)

	if derr != nil {
		return nil, fmt.Errorf("Decode error:", derr)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Err api:", derr)
	}

	return &responseData.DeviceId, nil
}

func isBackOfficeUser(userGroups []string) bool {

	if len(userGroups) > 0 {
		for _, group := range uams.AnyAdminScope {
			for _, userGroup := range userGroups {
				if string(group) == userGroup {
					return true
				}
			}
		}
	}
	return false
}

func verifyUserGroups(scopes []*string, userGroups []string) bool {
	for _, group := range scopes {
		for _, userGroup := range userGroups {
			if *group == userGroup {
				return true
			}
		}
	}

	return false
}

func convertActionsToPointers(actions []Action) []*string {
	var result []*string
	for _, action := range actions {
		actionStr := string(action)
		result = append(result, &actionStr)
	}
	return result
}
