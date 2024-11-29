package delivery

import auth "github.com/dwiw96/ran-user-management/internal/features/users"

type signupResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func toSignUpResponse(input *auth.User) signupResponse {
	return signupResponse{
		Username: input.Username,
		Email:    input.Email,
	}
}

type loginResponse struct {
	ID           int32  `json:"id"`
	Username     string `json:"Username"`
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func toLoginResponse(input *auth.User, accessToken, refreshToken string) loginResponse {
	return loginResponse{
		Username:     input.Username,
		Email:        input.Email,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

type refreshTokenResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}
