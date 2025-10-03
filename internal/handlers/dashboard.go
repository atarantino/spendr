package handlers

import (
	"net/http"
	"strconv"

	"spendr/cmd/web"
	"spendr/internal/auth"
	"spendr/internal/database"
	sqlc "spendr/internal/database/sqlc"

	"github.com/a-h/templ"
)

type DashboardHandler struct {
	db database.Service
}

func NewDashboardHandler(db database.Service) *DashboardHandler {
	return &DashboardHandler{
		db: db,
	}
}

func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's Plaid accounts (optional, for display)
	plaidItems, err := h.db.GetQueries().GetPlaidItemsByUserID(r.Context(), int32(userID))
	hasConnectedAccounts := err == nil && len(plaidItems) > 0

	// Get pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	const pageSize = 10
	offset := (page - 1) * pageSize

	// Get transactions for this user with pagination
	transactions := []interface{}{}
	totalPages := 0
	totalCount := int64(0)

	if hasConnectedAccounts {
		// Get total count
		totalCount, err = h.db.GetQueries().CountTransactionsByUserID(r.Context(), int32(userID))
		if err == nil && totalCount > 0 {
			totalPages = int((totalCount + int64(pageSize) - 1) / int64(pageSize))
		}

		// Get paginated transactions
		dbTransactions, err := h.db.GetQueries().GetTransactionsByUserIDPaginated(r.Context(), sqlc.GetTransactionsByUserIDPaginatedParams{
			UserID: int32(userID),
			Limit:  pageSize,
			Offset: int32(offset),
		})
		if err == nil {
			// Convert to interface slice for template
			for _, tx := range dbTransactions {
				transactions = append(transactions, tx)
			}
		}
	}

	templ.Handler(web.DashboardPage(userID, hasConnectedAccounts, transactions, page, totalPages)).ServeHTTP(w, r)
}
