package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/naveenkanuri/simplebank/db/sqlc"
)


type transferAccountRequest struct {
	FromAccountID int64 `json:"from_account_id" binding:"required,min=1"`
	ToAccountID int64 `json:"to_account_id" binding:"required,min=1"`
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

  amount, err := server.handleCurrency(ctx, req.FromAccountID, req.ToAccountID, req.Amount)
  if err != nil {
    return
  }

	arg := db.TransferTxParams{
    FromAccountID: req.FromAccountID,
    ToAccountID: req.ToAccountID,
    Amount: amount,
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

func (server *Server) handleCurrency(ctx *gin.Context, fromAccountID int64, toAccountID int64, amountToTransfer int64) (amount int64, err error) {
  fromAccount, err := server.store.GetAccount(ctx, fromAccountID)
  if err != nil {
    if err == sql.ErrNoRows {
      ctx.JSON(http.StatusNotFound, errorResponse(err))
      return
    }
    ctx.JSON(http.StatusInternalServerError, errorResponse(err))
    return
  }

  toAccount, err := server.store.GetAccount(ctx, toAccountID)
  if err != nil {
    if err == sql.ErrNoRows {
      ctx.JSON(http.StatusNotFound, errorResponse(err))
      return
    }
    ctx.JSON(http.StatusInternalServerError, errorResponse(err))
    return
  }

  log.Printf("From Account Currency: %s", fromAccount.Currency)
  log.Printf("To Account Currency: %s", toAccount.Currency)

  if fromAccount.Currency != toAccount.Currency {
    log.Printf("Converting currency from %s to %s", fromAccount.Currency, toAccount.Currency)
    amount, err = convertToUSD(amountToTransfer, fromAccount.Currency)
    log.Printf("Converted amount to USD: %d", amount)
    if err != nil {
      ctx.JSON(http.StatusInternalServerError, errorResponse(err))
      return
    }
    amount, err = convertFromUSD(amount, toAccount.Currency)
    log.Printf("Converted amount from USD: %d", amount)
    if err != nil {
      ctx.JSON(http.StatusInternalServerError, errorResponse(err))
      return
    }
  } else {
    amount = amountToTransfer
  }
  return
}
