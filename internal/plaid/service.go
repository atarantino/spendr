package plaid

import (
	"context"
	"fmt"
	"os"

	"github.com/plaid/plaid-go/v39/plaid"
)

type Service struct {
	client *plaid.APIClient
	env    plaid.Environment
}

func NewService() *Service {
	clientID := os.Getenv("PLAID_CLIENT_ID")
	secret := os.Getenv("PLAID_SECRET")
	env := os.Getenv("PLAID_ENV")

	if clientID == "" || secret == "" || env == "" {
		panic("PLAID_CLIENT_ID, PLAID_SECRET, and PLAID_ENV must be set")
	}

	var plaidEnv plaid.Environment
	switch env {
	case "sandbox":
		plaidEnv = plaid.Sandbox
	case "production":
		plaidEnv = plaid.Production
	default:
		panic(fmt.Sprintf("invalid PLAID_ENV: %s (must be sandbox or production)", env))
	}

	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	configuration.AddDefaultHeader("PLAID-SECRET", secret)
	configuration.UseEnvironment(plaidEnv)

	return &Service{
		client: plaid.NewAPIClient(configuration),
		env:    plaidEnv,
	}
}

type LinkTokenResponse struct {
	LinkToken  string `json:"link_token"`
	Expiration string `json:"expiration"`
}

func (s *Service) CreateLinkToken(ctx context.Context, userID int, redirectURI string) (*LinkTokenResponse, error) {
	clientUserID := fmt.Sprintf("%d", userID)
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: clientUserID,
	}

	request := plaid.NewLinkTokenCreateRequest(
		"Spendr",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
	)
	request.SetUser(user)
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})

	if redirectURI != "" {
		request.SetRedirectUri(redirectURI)
	}

	resp, _, err := s.client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create link token: %w", err)
	}

	expiration := resp.GetExpiration().Format("2006-01-02T15:04:05Z")
	return &LinkTokenResponse{
		LinkToken:  resp.GetLinkToken(),
		Expiration: expiration,
	}, nil
}

type ExchangeTokenResponse struct {
	AccessToken string `json:"access_token"`
	ItemID      string `json:"item_id"`
}

func (s *Service) ExchangePublicToken(ctx context.Context, publicToken string) (*ExchangeTokenResponse, error) {
	request := plaid.NewItemPublicTokenExchangeRequest(publicToken)

	resp, _, err := s.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*request).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to exchange public token: %w", err)
	}

	return &ExchangeTokenResponse{
		AccessToken: resp.GetAccessToken(),
		ItemID:      resp.GetItemId(),
	}, nil
}

type Account struct {
	AccountID    string  `json:"account_id"`
	Name         string  `json:"name"`
	OfficialName *string `json:"official_name,omitempty"`
	Type         string  `json:"type"`
	Subtype      *string `json:"subtype,omitempty"`
}

func (s *Service) GetAccounts(ctx context.Context, accessToken string) ([]Account, error) {
	request := plaid.NewAccountsGetRequest(accessToken)

	resp, _, err := s.client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(*request).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	accounts := make([]Account, 0, len(resp.GetAccounts()))
	for _, acc := range resp.GetAccounts() {
		account := Account{
			AccountID: acc.GetAccountId(),
			Name:      acc.GetName(),
			Type:      string(acc.GetType()),
		}

		if officialName := acc.GetOfficialName(); officialName != "" {
			account.OfficialName = &officialName
		}

		if subtype := acc.GetSubtype(); subtype != "" {
			subtypeStr := string(subtype)
			account.Subtype = &subtypeStr
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

type Transaction struct {
	TransactionID          string                  `json:"transaction_id"`
	AccountID              string                  `json:"account_id"`
	Amount                 float64                 `json:"amount"`
	Date                   string                  `json:"date"`
	AuthorizedDate         *string                 `json:"authorized_date,omitempty"`
	Name                   string                  `json:"name"`
	MerchantName           *string                 `json:"merchant_name,omitempty"`
	Pending                bool                    `json:"pending"`
	PaymentChannel         string                  `json:"payment_channel"`
	TransactionCode        *string                 `json:"transaction_code,omitempty"`
	ISOCurrencyCode        *string                 `json:"iso_currency_code,omitempty"`
	UnofficialCurrencyCode *string                 `json:"unofficial_currency_code,omitempty"`
	Location               map[string]interface{}  `json:"location,omitempty"`
	PaymentMeta            map[string]interface{}  `json:"payment_meta,omitempty"`
	PersonalFinanceCategory map[string]interface{} `json:"personal_finance_category,omitempty"`
	Counterparties         []map[string]interface{} `json:"counterparties,omitempty"`
}

type SyncResult struct {
	Added    []Transaction `json:"added"`
	Modified []Transaction `json:"modified"`
	Removed  []string      `json:"removed"`
	Cursor   string        `json:"next_cursor"`
	HasMore  bool          `json:"has_more"`
}

func (s *Service) SyncTransactions(ctx context.Context, accessToken string, cursor *string) (*SyncResult, error) {
	request := plaid.NewTransactionsSyncRequest(accessToken)
	if cursor != nil && *cursor != "" {
		request.SetCursor(*cursor)
	}

	resp, _, err := s.client.PlaidApi.TransactionsSync(ctx).TransactionsSyncRequest(*request).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to sync transactions: %w", err)
	}

	result := &SyncResult{
		Added:    make([]Transaction, 0),
		Modified: make([]Transaction, 0),
		Removed:  make([]string, 0),
		Cursor:   resp.GetNextCursor(),
		HasMore:  resp.GetHasMore(),
	}

	for _, tx := range resp.GetAdded() {
		result.Added = append(result.Added, convertTransaction(tx))
	}

	for _, tx := range resp.GetModified() {
		result.Modified = append(result.Modified, convertTransaction(tx))
	}

	for _, txID := range resp.GetRemoved() {
		result.Removed = append(result.Removed, txID.GetTransactionId())
	}

	return result, nil
}

func convertTransaction(tx plaid.Transaction) Transaction {
	t := Transaction{
		TransactionID:  tx.GetTransactionId(),
		AccountID:      tx.GetAccountId(),
		Amount:         tx.GetAmount(),
		Date:           tx.GetDate(),
		Name:           tx.GetName(),
		Pending:        tx.GetPending(),
		PaymentChannel: string(tx.GetPaymentChannel()),
	}

	if authDate := tx.GetAuthorizedDate(); authDate != "" {
		t.AuthorizedDate = &authDate
	}

	if merchantName := tx.GetMerchantName(); merchantName != "" {
		t.MerchantName = &merchantName
	}

	if txCode := tx.GetTransactionCode(); txCode != "" {
		txCodeStr := string(txCode)
		t.TransactionCode = &txCodeStr
	}

	if isoCurrency := tx.GetIsoCurrencyCode(); isoCurrency != "" {
		t.ISOCurrencyCode = &isoCurrency
	}

	if unofficialCurrency := tx.GetUnofficialCurrencyCode(); unofficialCurrency != "" {
		t.UnofficialCurrencyCode = &unofficialCurrency
	}

	if location := tx.GetLocation(); location.GetAddress() != "" || location.GetCity() != "" {
		t.Location = map[string]interface{}{
			"address":     location.GetAddress(),
			"city":        location.GetCity(),
			"region":      location.GetRegion(),
			"postal_code": location.GetPostalCode(),
			"country":     location.GetCountry(),
			"lat":         location.GetLat(),
			"lon":         location.GetLon(),
		}
	}

	if paymentMeta := tx.GetPaymentMeta(); paymentMeta.GetReferenceNumber() != "" {
		t.PaymentMeta = map[string]interface{}{
			"reference_number": paymentMeta.GetReferenceNumber(),
			"ppd_id":          paymentMeta.GetPpdId(),
			"payee":           paymentMeta.GetPayee(),
			"by_order_of":     paymentMeta.GetByOrderOf(),
			"payer":           paymentMeta.GetPayer(),
			"payment_method":  paymentMeta.GetPaymentMethod(),
			"payment_processor": paymentMeta.GetPaymentProcessor(),
			"reason":          paymentMeta.GetReason(),
		}
	}

	if pfc := tx.GetPersonalFinanceCategory(); pfc.GetPrimary() != "" {
		t.PersonalFinanceCategory = map[string]interface{}{
			"primary":   pfc.GetPrimary(),
			"detailed":  pfc.GetDetailed(),
		}
	}

	if counterparties := tx.GetCounterparties(); len(counterparties) > 0 {
		t.Counterparties = make([]map[string]interface{}, 0, len(counterparties))
		for _, cp := range counterparties {
			t.Counterparties = append(t.Counterparties, map[string]interface{}{
				"name":           cp.GetName(),
				"type":           cp.GetType(),
				"logo_url":       cp.GetLogoUrl(),
				"website":        cp.GetWebsite(),
				"entity_id":      cp.GetEntityId(),
				"confidence_level": cp.GetConfidenceLevel(),
			})
		}
	}

	return t
}

func (s *Service) GetInstitutionName(ctx context.Context, institutionID string) (string, error) {
	countryCode := plaid.COUNTRYCODE_US
	request := plaid.NewInstitutionsGetByIdRequest(institutionID, []plaid.CountryCode{countryCode})

	resp, _, err := s.client.PlaidApi.InstitutionsGetById(ctx).InstitutionsGetByIdRequest(*request).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get institution: %w", err)
	}

	institution := resp.GetInstitution()
	return institution.Name, nil
}
