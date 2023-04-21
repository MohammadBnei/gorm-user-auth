package model

type UserCreateDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserUpdateDTO struct {
	Email string `json:"email"`
}
