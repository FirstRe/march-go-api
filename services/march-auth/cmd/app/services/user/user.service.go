package userService

import (
	"core/app/common/jwt"
	"fmt"
	"log"

	"go-graphql/cmd/app/common"
	gormDb "go-graphql/cmd/app/common/gorm"
	"go-graphql/cmd/app/graph/model"
	"go-graphql/cmd/app/graph/types"

	"golang.org/x/crypto/bcrypt"
)

func CreateUser(input *types.UserInputParams) (*types.ResponseCreateUser, error) {
	findDup := model.User{}
	gormDb.Repos.Model(&model.User{}).Where("name = ? OR email = ?", input.Name, input.Email).Find(&findDup)

	if findDup != (model.User{}) {
		reponseError := types.ResponseCreateUser{
			Status: common.StatusResponse(400, "Duplicated"),
			Data:   nil,
		}
		return &reponseError, nil
	}

	password := []byte(input.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	userCreate := model.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	re := gormDb.Repos.Model(&model.User{}).Create(&userCreate)

	log.Printf("result: %+v\n", re.Error)
	log.Printf("userCreate: %+v\n", userCreate.CreatedAt)

	if re.Error != nil {
		reponseError := types.ResponseCreateUser{
			Status: common.StatusResponse(500, ""),
			Data:   nil,
		}
		return &reponseError, nil
	}

	userResponse := types.User{
		ID: userCreate.ID,
		Name:      userCreate.Name,
		Email:     userCreate.Email,
		CreatedAt: userCreate.CreatedAt,
		UpdatedAt: userCreate.UpdatedAt.String(),
	}

	response := types.ResponseCreateUser{
		Status: common.StatusResponse(200, ""),
		Data:   &userResponse,
	}

	log.Printf("response: %+v\n", response.Data)
	return &response, nil
}

func Price() (*types.Price, error) {

	price := &types.Price{
		Price: 12.123,
	}
	formattedPrice := fmt.Sprintf("%.2f", 12.123)

	fmt.Println(formattedPrice)
	return price, nil
	
}

func Login(input *types.LoginInputParams) (*types.ResponseLogin, error) {
	findUser := model.User{}
	reponseError := types.ResponseLogin{}
	gormDb.Repos.Model(&model.User{}).Where("name = ? OR email = ?", input.Username, input.Username).Find(&findUser)

	if findUser == (model.User{}) {
		reponseError = types.ResponseLogin{
			Status: common.StatusResponse(401, ""),
			Data:   nil,
		}
		return &reponseError, nil
	}

	errCompare := bcrypt.CompareHashAndPassword([]byte(findUser.Password), []byte(input.Password))
	if errCompare != nil {
		reponseError = types.ResponseLogin{
			Status: common.StatusResponse(401, ""),
			Data:   nil,
		}
		return &reponseError, nil
	}

	token, errToken := jwt.JwtGenerate(findUser.ID)

	if errToken != nil {
		reponseError = types.ResponseLogin{
			Status: common.StatusResponse(500, ""),
			Data:   nil,
		}
		return &reponseError, nil
	}

	tokenData := types.Token{
		AccessToken:  token,
		RefreshToken: &token,
		Username:     &findUser.Email,
		UserID:       &findUser.ID,
	}

	reponseData := types.ResponseLogin{
		Status: common.StatusResponse(200, ""),
		Data:   &tokenData,
	}

	return &reponseData, nil

}
