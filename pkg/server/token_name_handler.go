package server

import (
	"github.com/OdysseyMomentumExperience/token-service/pkg/constants"
	"github.com/OdysseyMomentumExperience/token-service/pkg/types"
	"github.com/OdysseyMomentumExperience/token-service/pkg/web3/eth"
	"github.com/OdysseyMomentumExperience/token-service/pkg/xhttp"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"net/http"
)

type TokenNameHandler struct {
	service   *eth.EthNameService
	validator *validator.Validate
}

func NewTokenNameHandler(service *eth.EthNameService) *TokenNameHandler {
	return &TokenNameHandler{service: service, validator: validator.New()}
}

func (t *TokenNameHandler) NameHandler(w http.ResponseWriter, r *http.Request) {
	req := new(types.NameRequest)

	if err := render.DecodeJSON(r.Body, req); err != nil {
		err := errors.Wrap(err, "failed to decode body")
		http.Error(w, err.Error(), 400)
		return
	}

	err := t.validator.Struct(req)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		xhttp.Error(w, validationErrors, 400)
		return
	}

	tokenName, err := t.service.GetTokenName(req)
	if err != nil {
		xhttp.Error(w, err, 400)
		return
	}
	if tokenName == "" {
		tokenName = constants.UNKNOWN
	}
	render.Respond(w, r, tokenName)
}
