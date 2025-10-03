package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"spendr/internal/auth"
	"spendr/internal/database"
	sqlc "spendr/internal/database/sqlc"
	"spendr/internal/plaid"

	"github.com/jackc/pgx/v5/pgtype"
)

type PlaidHandler struct {
	plaidService *plaid.Service
	db           database.Service
}

func NewPlaidHandler(plaidService *plaid.Service, db database.Service) *PlaidHandler {
	return &PlaidHandler{
		plaidService: plaidService,
		db:           db,
	}
}

type CreateLinkTokenRequest struct {
	RedirectURI string `json:"redirect_uri,omitempty"`
}

func (h *PlaidHandler) CreateLinkToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateLinkTokenRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	linkToken, err := h.plaidService.CreateLinkToken(r.Context(), userID, req.RedirectURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create link token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(linkToken)
}

type ExchangePublicTokenRequest struct {
	PublicToken   string `json:"public_token"`
	InstitutionID string `json:"institution_id"`
}

func (h *PlaidHandler) ExchangePublicToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ExchangePublicTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Exchange public token for access token
	exchangeResp, err := h.plaidService.ExchangePublicToken(r.Context(), req.PublicToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
		return
	}

	// Get institution name
	institutionName, err := h.plaidService.GetInstitutionName(r.Context(), req.InstitutionID)
	if err != nil {
		// Don't fail if we can't get institution name
		institutionName = "Unknown"
	}

	// Store Plaid item in database
	plaidItem, err := h.db.GetQueries().CreatePlaidItem(r.Context(), sqlc.CreatePlaidItemParams{
		UserID:          int32(userID),
		AccessToken:     exchangeResp.AccessToken,
		ItemID:          exchangeResp.ItemID,
		InstitutionName: pgtype.Text{String: institutionName, Valid: true},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to store Plaid item: %v", err), http.StatusInternalServerError)
		return
	}

	// Get accounts for this item
	accounts, err := h.plaidService.GetAccounts(r.Context(), exchangeResp.AccessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get accounts: %v", err), http.StatusInternalServerError)
		return
	}

	// Store accounts in database
	for _, account := range accounts {
		var officialName pgtype.Text
		if account.OfficialName != nil {
			officialName = pgtype.Text{String: *account.OfficialName, Valid: true}
		}

		var subtype pgtype.Text
		if account.Subtype != nil {
			subtype = pgtype.Text{String: *account.Subtype, Valid: true}
		}

		_, err := h.db.GetQueries().CreatePlaidAccount(r.Context(), sqlc.CreatePlaidAccountParams{
			PlaidItemID:  plaidItem.ID,
			AccountID:    account.AccountID,
			Name:         account.Name,
			OfficialName: officialName,
			Type:         account.Type,
			Subtype:      subtype,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to store account: %v", err), http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"success":     true,
		"item_id":     exchangeResp.ItemID,
		"institution": institutionName,
		"accounts":    len(accounts),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PlaidHandler) SyncTransactions(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all Plaid items for this user
	plaidItems, err := h.db.GetQueries().GetPlaidItemsByUserID(r.Context(), int32(userID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get Plaid items: %v", err), http.StatusInternalServerError)
		return
	}

	totalAdded := 0
	totalModified := 0
	totalRemoved := 0

	// Sync transactions for each item
	for _, item := range plaidItems {
		result, err := h.syncItemTransactions(r.Context(), item, userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to sync transactions for item %s: %v", item.ItemID, err), http.StatusInternalServerError)
			return
		}
		totalAdded += result.Added
		totalModified += result.Modified
		totalRemoved += result.Removed
	}

	response := map[string]interface{}{
		"success":               true,
		"items_synced":          len(plaidItems),
		"transactions_added":    totalAdded,
		"transactions_modified": totalModified,
		"transactions_removed":  totalRemoved,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type syncItemResult struct {
	Added    int
	Modified int
	Removed  int
}

func (h *PlaidHandler) syncItemTransactions(ctx context.Context, item sqlc.GetPlaidItemsByUserIDRow, userID int) (*syncItemResult, error) {
	result := &syncItemResult{}

	var cursor *string
	if item.TransactionsCursor.Valid {
		cursor = &item.TransactionsCursor.String
	}

	// Get accounts for this item to map account_id to plaid_account_id
	accounts, err := h.db.GetQueries().GetPlaidAccountsByItemID(ctx, item.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	accountMap := make(map[string]int32)
	for _, acc := range accounts {
		accountMap[acc.AccountID] = acc.ID
	}

	hasMore := true
	for hasMore {
		syncResult, err := h.plaidService.SyncTransactions(ctx, item.AccessToken, cursor)
		if err != nil {
			return nil, fmt.Errorf("failed to sync transactions: %w", err)
		}

		// Process added transactions
		for _, tx := range syncResult.Added {
			plaidAccountID, ok := accountMap[tx.AccountID]
			if !ok {
				continue // Skip if account not found
			}

			if err := h.createTransaction(ctx, tx, plaidAccountID, userID); err != nil {
				// Log but don't fail on duplicate transactions
				if err.Error() != "no rows returned" {
					return nil, fmt.Errorf("failed to create transaction: %w", err)
				}
			} else {
				result.Added++
			}
		}

		// Process modified transactions (update pending status)
		for _, tx := range syncResult.Modified {
			if err := h.db.GetQueries().UpdateTransactionPendingStatus(ctx, sqlc.UpdateTransactionPendingStatusParams{
				TransactionID: tx.TransactionID,
				Pending:       tx.Pending,
			}); err != nil {
				return nil, fmt.Errorf("failed to update transaction: %w", err)
			}
			result.Modified++
		}

		result.Removed += len(syncResult.Removed)

		// Update cursor
		_, err = h.db.GetQueries().UpdatePlaidItemCursor(ctx, sqlc.UpdatePlaidItemCursorParams{
			ItemID:             item.ItemID,
			TransactionsCursor: pgtype.Text{String: syncResult.Cursor, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update cursor: %w", err)
		}

		cursor = &syncResult.Cursor
		hasMore = syncResult.HasMore
	}

	return result, nil
}

func (h *PlaidHandler) createTransaction(ctx context.Context, tx plaid.Transaction, plaidAccountID int32, userID int) error {
	var authorizedDate pgtype.Date
	if tx.AuthorizedDate != nil {
		authorizedDate = pgtype.Date{Time: parseDate(*tx.AuthorizedDate), Valid: true}
	}

	var merchantName pgtype.Text
	if tx.MerchantName != nil {
		merchantName = pgtype.Text{String: *tx.MerchantName, Valid: true}
	}

	var transactionCode pgtype.Text
	if tx.TransactionCode != nil {
		transactionCode = pgtype.Text{String: *tx.TransactionCode, Valid: true}
	}

	var isoCurrencyCode pgtype.Text
	if tx.ISOCurrencyCode != nil {
		isoCurrencyCode = pgtype.Text{String: *tx.ISOCurrencyCode, Valid: true}
	}

	var unofficialCurrencyCode pgtype.Text
	if tx.UnofficialCurrencyCode != nil {
		unofficialCurrencyCode = pgtype.Text{String: *tx.UnofficialCurrencyCode, Valid: true}
	}

	location, _ := json.Marshal(tx.Location)
	paymentMeta, _ := json.Marshal(tx.PaymentMeta)
	personalFinanceCategory, _ := json.Marshal(tx.PersonalFinanceCategory)
	counterparties, _ := json.Marshal(tx.Counterparties)

	amount := pgtype.Numeric{}
	amount.Scan(fmt.Sprintf("%.2f", tx.Amount))

	_, err := h.db.GetQueries().CreateTransaction(ctx, sqlc.CreateTransactionParams{
		UserID:                  int32(userID),
		PlaidAccountID:          plaidAccountID,
		TransactionID:           tx.TransactionID,
		AccountID:               tx.AccountID,
		Amount:                  amount,
		Date:                    pgtype.Date{Time: parseDate(tx.Date), Valid: true},
		AuthorizedDate:          authorizedDate,
		Name:                    tx.Name,
		MerchantName:            merchantName,
		Pending:                 tx.Pending,
		PaymentChannel:          tx.PaymentChannel,
		TransactionCode:         transactionCode,
		IsoCurrencyCode:         isoCurrencyCode,
		UnofficialCurrencyCode:  unofficialCurrencyCode,
		Location:                location,
		PaymentMeta:             paymentMeta,
		PersonalFinanceCategory: personalFinanceCategory,
		Counterparties:          counterparties,
	})

	return err
}

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}

func (h *PlaidHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	plaidItems, err := h.db.GetQueries().GetPlaidItemsByUserID(r.Context(), int32(userID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get Plaid items: %v", err), http.StatusInternalServerError)
		return
	}

	type AccountWithInstitution struct {
		sqlc.PlaidAccount
		InstitutionName string `json:"institution_name"`
	}

	accounts := make([]AccountWithInstitution, 0)

	for _, item := range plaidItems {
		itemAccounts, err := h.db.GetQueries().GetPlaidAccountsByItemID(r.Context(), item.ID)
		if err != nil {
			continue
		}

		institutionName := "Unknown"
		if item.InstitutionName.Valid {
			institutionName = item.InstitutionName.String
		}

		for _, acc := range itemAccounts {
			accounts = append(accounts, AccountWithInstitution{
				PlaidAccount:    acc,
				InstitutionName: institutionName,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}
