package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"ukiran.com/snippetbox/internal/models"
	"ukiran.com/snippetbox/internal/validator"
	"ukiran.com/snippetbox/ui/html/pages"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
	}

	data := app.NewTemplateData(r)
	data.Snippets = snippets

	err = pages.HomePage(data).Render(r.Context(), w)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.NewTemplateData(r)
	err := pages.AboutPage(data).Render(r.Context(), w)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.NewTemplateData(r)
	data.Snippet = snippet

	err = pages.ViewPage(data).Render(r.Context(), w)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) snippetDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	err = app.snippets.Delete(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	// add Flash msg to session
	app.sessionManager.Put(
		r.Context(), "flash", fmt.Sprintf("Snippet %d deleted!", id),
	)
	// redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.NewTemplateData(r)
	data.Form = models.SnippetCreateForm{
		Expires: 365,
	}

	err := pages.CreatePage(data).Render(r.Context(), w)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) snippetCreatePost(
	w http.ResponseWriter, r *http.Request,
) {
	var form models.SnippetCreateForm

	err := app.decodePostForm(&form, r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title),
		"title", "This field cannot be blank")
	form.CheckField(
		validator.MaxChars(form.Title, 100),
		"title",
		"This field cannot be more than 100 characters long",
	)
	form.CheckField(validator.NotBlank(form.Content),
		"content", "This field cannot be blank")
	form.CheckField(
		validator.PermittedValue(form.Expires, 1, 7, 365),
		"expires",
		"This field must equal 1, 7 or 365",
	)

	if !form.Valid() {
		data := app.NewTemplateData(r)
		data.Form = form

		w.WriteHeader(http.StatusUnprocessableEntity)

		err := pages.CreatePage(data).Render(r.Context(), w)
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (app *application) accountView(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	user, err := app.users.Get(userId)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	data := app.NewTemplateData(r)
	data.User = user

	err = pages.AccountPage(data).Render(r.Context(), w)
	if err != nil {
		app.serverError(w, r, err)
	}
	return
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.NewTemplateData(r)
	data.Form = models.UserSignupForm{}
	err := pages.SignupPage(data).Render(r.Context(), w)
	if err != nil {
		app.serverError(w, r, err)
	}
	return
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form models.UserSignupForm

	err := app.decodePostForm(&form, r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name),
		"name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email),
		"email", "This field cannot be blank")
	form.CheckField(
		validator.Matches(form.Email, validator.EmailRX),
		"email",
		"This field must be a valid email address",
	)
	form.CheckField(validator.NotBlank(form.Password),
		"password", "This field cannot be blank")
	form.CheckField(
		validator.MinChars(form.Password, 8),
		"password",
		"This field must be at least 8 characters long",
	)

	if !form.Valid() {
		data := app.NewTemplateData(r)
		data.Form = form
		w.WriteHeader(http.StatusUnprocessableEntity)
		err := pages.SignupPage(data).Render(r.Context(), w)
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := app.NewTemplateData(r)
			data.Form = form
			w.WriteHeader(http.StatusUnprocessableEntity)
			err = pages.SignupPage(data).Render(r.Context(), w)
			if err != nil {
				app.serverError(w, r, err)
			}
			return
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	app.sessionManager.Put(r.Context(),
		"flash", "Your signup was successful. Please log in")

	// And redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.NewTemplateData(r)
	data.Form = models.UserLoginForm{}
	err := pages.LoginPage(data).Render(r.Context(), w)
	if err != nil {
		app.serverError(w, r, err)
	}
	return
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form models.UserLoginForm
	err := app.decodePostForm(&form, r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.Email),
		"email", "This field cannot be blank")
	form.CheckField(
		validator.Matches(form.Email, validator.EmailRX),
		"email",
		"This field must be a valid email address",
	)
	form.CheckField(validator.NotBlank(form.Password),
		"password", "This field cannot be blank")

	if !form.Valid() {
		data := app.NewTemplateData(r)
		data.Form = form
		err := pages.LoginPage(data).Render(r.Context(), w)
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or Password is incorrect")
			data := app.NewTemplateData(r)
			data.Form = form
			err = pages.LoginPage(data).Render(r.Context(), w)
			if err != nil {
				app.serverError(w, r, err)
			}
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	path := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")
	if path != "" {
		http.Redirect(w, r, path, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(),
		"flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) accountPasswordUpdate(
	w http.ResponseWriter, r *http.Request,
) {
	data := app.NewTemplateData(r)
	data.Form = models.AccountPasswordUpdateForm{}
	err := pages.AccountPwdPage(data).Render(r.Context(), w)
	if err != nil {
		app.serverError(w, r, err)
	}
	return
}

func (app *application) accountPasswordUpdatePost(
	w http.ResponseWriter, r *http.Request,
) {
	var form models.AccountPasswordUpdateForm

	err := app.decodePostForm(&form, r)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.CurrentPassword),
		"current_password", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPassword),
		"new_password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 8),
		"new_password", "This field must be at least 8 characters long")
	form.CheckField(validator.NotBlank(form.NewPasswordConfirm),
		"new_password_confirm", "This field cannot be blank")
	form.CheckField(validator.StringEquals(
		form.NewPassword, form.NewPasswordConfirm),
		"new_password_confirm", "Passwords do not match")

	if !form.Valid() {
		data := app.NewTemplateData(r)
		data.Form = form
		w.WriteHeader(http.StatusUnprocessableEntity)
		err := pages.AccountPwdPage(data).Render(r.Context(), w)
		if err != nil {
			app.serverError(w, r, err)
		}
		return
	}
	userId := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	err = app.users.PasswordUpdate(userId, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddFieldError("current_password", "Current password is incorrect")

			data := app.NewTemplateData(r)
			data.Form = form

			err := pages.AccountPwdPage(data).Render(r.Context(), w)
			if err != nil {
				app.serverError(w, r, err)
			}
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Your password has been updated!")
	http.Redirect(w, r, "/account/view", http.StatusSeeOther)
}
