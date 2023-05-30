package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/naveenkanuri/simplebank/db/sqlc"
	"github.com/naveenkanuri/simplebank/token"
)

type transferAccountRequest struct {
	FromAccountID int64 `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64 `json:"to_account_id" binding:"required,min=1"`
	Amount        int64 `json:"amount" binding:"required,gt=0"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, valid := server.validAccount(ctx, req.FromAccountID)
	if !valid {
		return
	}
	authPayload := ctx.MustGet(authoriztionPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	toAccount, valid := server.validAccount(ctx, req.ToAccountID)
	if !valid {
		return
	}

	amount, err := server.handleCurrency(ctx, fromAccount, toAccount, req.Amount)
	if err != nil {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func convertToUSD(amount int64, currency string) (int64, error) {
	floatAmount := float64(amount)
	switch currency {
	case "USD":
		return int64(floatAmount), nil
	case "EUR":
		return int64(floatAmount * 0.9), nil
	case "CAD":
		return int64(floatAmount * 0.7), nil
	case "INR":
		return int64(floatAmount * 0.0125), nil
	default:
		return 0, nil
	}
}

func convertFromUSD(amount int64, currency string) (int64, error) {
	floatAmount := float64(amount)
	switch currency {
	case "USD":
		return int64(floatAmount), nil
	case "EUR":
		return int64(floatAmount / 0.9), nil
	case "CAD":
		return int64(floatAmount / 0.7), nil
	case "INR":
		return int64(floatAmount / 0.0125), nil
	default:
		return 0, nil
	}
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}
	return account, true
}

func (server *Server) handleCurrency(ctx *gin.Context, fromAccount db.Account, toAccount db.Account, amountToTransfer int64) (amount int64, err error) {
	if fromAccount.Currency != toAccount.Currency {
		amount, err = convertToUSD(amountToTransfer, fromAccount.Currency)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		amount, err = convertFromUSD(amount, toAccount.Currency)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	} else {
		amount = amountToTransfer
	}
	return
}
