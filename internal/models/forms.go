package models

import "ukiran.com/snippetbox/internal/validator"

type SnippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

type UserSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type UserLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type AccountPasswordUpdateForm struct {
	CurrentPassword     string `form:"current_password"`
	NewPassword         string `form:"new_password"`
	NewPasswordConfirm  string `form:"new_password_confirm"`
	validator.Validator `form:"-"`
}
