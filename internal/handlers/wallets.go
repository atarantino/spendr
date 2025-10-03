package handlers

import (
	"fmt"
	"net/http"

	"spendr/cmd/web"
	"spendr/internal/auth"
	"spendr/internal/database"
	sqlc "spendr/internal/database/sqlc"

	"github.com/a-h/templ"
)

type WalletsHandler struct {
	db database.Service
}

func NewWalletsHandler(db database.Service) *WalletsHandler {
	return &WalletsHandler{
		db: db,
	}
}

func (h *WalletsHandler) WalletsPage(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's wallet
	wallet, err := h.db.GetQueries().GetWalletByUserID(r.Context(), int32(userID))
	hasWallet := err == nil

	var members []sqlc.GetWalletMembersByWalletIDRow
	if hasWallet {
		members, _ = h.db.GetQueries().GetWalletMembersByWalletID(r.Context(), wallet.ID)
	}

	var walletPtr *sqlc.Wallet
	if hasWallet {
		walletPtr = &wallet
	}

	templ.Handler(web.WalletsPage(userID, walletPtr, members, hasWallet)).ServeHTTP(w, r)
}

func (h *WalletsHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Wallet name is required", http.StatusBadRequest)
		return
	}

	// Create wallet
	wallet, err := h.db.GetQueries().CreateWallet(r.Context(), name)
	if err != nil {
		http.Error(w, "Failed to create wallet", http.StatusInternalServerError)
		return
	}

	// Add creator as first member
	err = h.db.GetQueries().AddWalletMember(r.Context(), sqlc.AddWalletMemberParams{
		WalletID: wallet.ID,
		UserID:   int32(userID),
	})
	if err != nil {
		http.Error(w, "Failed to add member to wallet", http.StatusInternalServerError)
		return
	}

	// Redirect to wallets page
	w.Header().Set("HX-Redirect", "/wallets")
	w.WriteHeader(http.StatusOK)
}

func (h *WalletsHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Get the user's wallet
	wallet, err := h.db.GetQueries().GetWalletByUserID(r.Context(), int32(userID))
	if err != nil {
		http.Error(w, "You don't have a wallet", http.StatusBadRequest)
		return
	}

	// Find user by email
	newUser, err := h.db.GetQueries().GetUserByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, "User not found with that email", http.StatusNotFound)
		return
	}

	// Add member to wallet
	err = h.db.GetQueries().AddWalletMember(r.Context(), sqlc.AddWalletMemberParams{
		WalletID: wallet.ID,
		UserID:   newUser.ID,
	})
	if err != nil {
		http.Error(w, "Failed to add member (they may already be in the wallet)", http.StatusInternalServerError)
		return
	}

	// Redirect to wallets page
	w.Header().Set("HX-Redirect", "/wallets")
	w.WriteHeader(http.StatusOK)
}

func (h *WalletsHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get walletID and memberID from URL params (chi will provide these)
	walletIDStr := r.PathValue("walletID")
	memberIDStr := r.PathValue("memberID")

	var walletID, memberID int32
	if _, err := fmt.Sscanf(walletIDStr, "%d", &walletID); err != nil {
		http.Error(w, "Invalid wallet ID", http.StatusBadRequest)
		return
	}
	if _, err := fmt.Sscanf(memberIDStr, "%d", &memberID); err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Verify user has access to this wallet
	wallet, err := h.db.GetQueries().GetWalletByUserID(r.Context(), int32(userID))
	if err != nil || wallet.ID != walletID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Remove member
	err = h.db.GetQueries().RemoveWalletMember(r.Context(), sqlc.RemoveWalletMemberParams{
		WalletID: walletID,
		UserID:   memberID,
	})
	if err != nil {
		http.Error(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
