package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/lozanov95/sss/internal/data"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorSecretNotFound = errors.New("secret not found")
)

func (app *application) extractSecretHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("id")

	secret, err := app.models.Secrets.Get(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "failed to find secret")
		return
	}

	if len(secret.Passphrase) != 0 {
		var input struct {
			Passphrase string `json:"passphrase"`
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "missing/invalid passphrase")
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(body, &input)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(err)
			fmt.Fprintf(w, "failed to parse passphrase")
			return
		}

		if !verifyPassword(secret.Passphrase, input.Passphrase) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "missing/invalid passphrase")
			return
		}
	}

	if err = app.models.Secrets.Delete(secret.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal server error")
		return
	}

	fmt.Fprintf(w, `{"secret":"%s"}`, secret.Value)
}

func (app *application) createSecretHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Secret     string `json:"secret"`
		Passphrase string `json:"passphrase"`
		ExpiresAt  string `json:"expiresAt"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to parse body")
		return
	}
	defer r.Body.Close()

	if input.Secret == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "missing value for key 'secret'")
		return
	}

	expiresAt, err := time.Parse(time.RFC3339, input.ExpiresAt)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.logger.Error("failed to parse date " + err.Error())
		fmt.Fprintf(w, "expiresAt must be in RFC3339 format (YYYY-MM-DDTHH:mm:ssZ00:00)")
		return
	}

	var encrPassphrase []byte
	if input.Passphrase != "" {
		encrPassphrase = hashAndSalt(input.Passphrase)
	}

	secret := data.Secret{
		Key:        rand.Text(),
		Value:      input.Secret,
		Passphrase: encrPassphrase,
		ExpiresAt:  expiresAt,
	}

	err = app.models.Secrets.Insert(&secret)
	if err != nil {
		app.logger.Error("failed to put secret " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to save secret")
		return
	}

	fmt.Fprintf(w, `{"secretId":"%s"}`, secret.Key)
}

func hashAndSalt(pwd string) []byte {
	bPass := []byte(pwd)
	hash, err := bcrypt.GenerateFromPassword(bPass, bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	}

	return hash
}

func verifyPassword(hashedPwd []byte, plainPwd string) bool {
	newPwdHash := []byte(plainPwd)
	if err := bcrypt.CompareHashAndPassword(hashedPwd, newPwdHash); err != nil {
		return false
	}

	return true
}
