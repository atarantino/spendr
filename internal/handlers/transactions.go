package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"spendr/internal/auth"
	"spendr/internal/database"
	sqlc "spendr/internal/database/sqlc"

	"github.com/go-chi/chi/v5"
)

var (
	errInvalidCategoryType     = errors.New("invalid category type")
	errTransactionNotFound     = errors.New("transaction not found")
	errUnauthorizedTransaction = errors.New("unauthorized transaction access")
)

type TransactionHandler struct {
	db database.Service
}

func NewTransactionHandler(db database.Service) *TransactionHandler {
	return &TransactionHandler{
		db: db,
	}
}

func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 20 // Default limit

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	// Get total count
	totalCount, err := h.db.GetQueries().CountTransactionsByUserID(r.Context(), int32(userID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to count transactions: %v", err), http.StatusInternalServerError)
		return
	}

	// Get paginated transactions
	transactions, err := h.db.GetQueries().GetTransactionsByUserIDPaginated(r.Context(), sqlc.GetTransactionsByUserIDPaginatedParams{
		UserID: int32(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get transactions: %v", err), http.StatusInternalServerError)
		return
	}

	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))

	response := map[string]interface{}{
		"transactions": transactions,
		"page":         page,
		"limit":        limit,
		"total_count":  totalCount,
		"total_pages":  totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TransactionHandler) GetUncategorizedTransactions(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	wallet, err := h.db.GetQueries().GetWalletByUserID(r.Context(), int32(userID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get wallet: %v", err), http.StatusInternalServerError)
		return
	}

	transactions, err := h.db.GetQueries().GetUncategorizedTransactionsByUserID(r.Context(), sqlc.GetUncategorizedTransactionsByUserIDParams{
		UserID:   int32(userID),
		WalletID: wallet.ID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get uncategorized transactions: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func (h *TransactionHandler) CategorizeTransaction(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	transactionIDStr := chi.URLParam(r, "id")
	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	walletIDStr := r.FormValue("wallet_id")
	walletID, err := strconv.Atoi(walletIDStr)
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}

	categoryType := r.FormValue("category_type")
	err = h.categorizeTransaction(r.Context(), int32(userID), int32(transactionID), int32(walletID), categoryType)
	if handled := handleCategorizationError(w, err); handled {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TransactionHandler) UncategorizeTransaction(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	transactionIDStr := chi.URLParam(r, "id")
	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	walletIDStr := chi.URLParam(r, "walletID")
	walletID, err := strconv.Atoi(walletIDStr)
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}

	err = h.db.GetQueries().DeleteTransactionCategorization(r.Context(), sqlc.DeleteTransactionCategorizationParams{
		TransactionID: int32(transactionID),
		WalletID:      int32(walletID),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to uncategorize transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TransactionHandler) GetSharedTransactions(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	walletIDStr := chi.URLParam(r, "walletID")
	walletID, err := strconv.Atoi(walletIDStr)
	if err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}

	transactions, err := h.db.GetQueries().GetSharedTransactionsByWalletID(r.Context(), int32(walletID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get shared transactions: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func (h *TransactionHandler) UncategorizedTransactionsPage(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *TransactionHandler) categorizeTransaction(ctx context.Context, userID, transactionID, walletID int32, categoryType string) error {
	if categoryType != "shared" && categoryType != "individual" {
		return errInvalidCategoryType
	}

	transaction, err := h.db.GetQueries().GetTransactionByID(ctx, transactionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errTransactionNotFound
		}

		return fmt.Errorf("get transaction: %w", err)
	}

	if transaction.UserID != userID {
		return errUnauthorizedTransaction
	}

	_, err = h.db.GetQueries().CreateTransactionCategorization(ctx, sqlc.CreateTransactionCategorizationParams{
		TransactionID:       transactionID,
		WalletID:            walletID,
		CategoryType:        categoryType,
		CategorizedByUserID: userID,
	})
	if err != nil {
		return fmt.Errorf("create transaction categorization: %w", err)
	}

	return nil
}

func handleCategorizationError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, errInvalidCategoryType):
		http.Error(w, "Invalid category type (must be 'shared' or 'individual')", http.StatusBadRequest)
	case errors.Is(err, errTransactionNotFound):
		http.Error(w, "Transaction not found", http.StatusNotFound)
	case errors.Is(err, errUnauthorizedTransaction):
		http.Error(w, "Unauthorized", http.StatusForbidden)
	default:
		http.Error(w, fmt.Sprintf("Failed to categorize transaction: %v", err), http.StatusInternalServerError)
	}

	return true
}
