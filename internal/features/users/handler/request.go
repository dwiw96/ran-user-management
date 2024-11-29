package delivery

import (
	auth "github.com/dwiw96/ran-user-management/internal/features/users"
)

type signupRequest struct {
	Username string `json:"username" validate:"required,min=2"`
	Email    string `json:"email" validate:"email,max=255"`
	Password string `json:"password" validate:"min=7,required_with=alphanum"`
}

func toSignUpRequest(input signupRequest) auth.SignupRequest {
	return auth.SignupRequest{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
	}
}

type signinRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=7"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}
