package account

import (
	"encoding/json"
	"fmt"

	"github.com/coltoneshaw/ynab.go/api"
)

// NewService facilitates the creation of a new account service instance
func NewService(c api.ClientReaderWriter) *Service {
	return &Service{c}
}

// Service wraps YNAB account API endpoints
type Service struct {
	c api.ClientReaderWriter
}

// GetAccounts fetches the list of accounts from a budget
// https://api.youneedabudget.com/v1#/Accounts/getAccounts
func (s *Service) GetAccounts(budgetID string, f *api.Filter) (*SearchResultSnapshot, error) {
	resModel := struct {
		Data struct {
			Accounts        []*Account `json:"accounts"`
			ServerKnowledge uint64     `json:"server_knowledge"`
		} `json:"data"`
	}{}

	url := fmt.Sprintf("/budgets/%s/accounts", budgetID)
	if f != nil {
		url = fmt.Sprintf("%s?%s", url, f.ToQuery())
	}
	if err := s.c.GET(url, &resModel); err != nil {
		return nil, err
	}

	return &SearchResultSnapshot{
		Accounts:        resModel.Data.Accounts,
		ServerKnowledge: resModel.Data.ServerKnowledge,
	}, nil
}

// GetAccount fetches a specific account from a budget
// https://api.youneedabudget.com/v1#/Accounts/getAccountById
func (s *Service) GetAccount(budgetID, accountID string) (*Account, error) {
	resModel := struct {
		Data struct {
			Account *Account `json:"account"`
		} `json:"data"`
	}{}

	url := fmt.Sprintf("/budgets/%s/accounts/%s", budgetID, accountID)
	if err := s.c.GET(url, &resModel); err != nil {
		return nil, err
	}
	return resModel.Data.Account, nil
}

// CreateAccount creates a new account in a budget
// https://api.youneedabudget.com/v1#/Accounts/createAccount
func (s *Service) CreateAccount(budgetID string, p PayloadAccount) (*Account, error) {
	payload := struct {
		Account *PayloadAccount `json:"account"`
	}{
		&p,
	}

	buf, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}

	resModel := struct {
		Data struct {
			Account *Account `json:"account"`
		} `json:"data"`
	}{}

	url := fmt.Sprintf("/budgets/%s/accounts", budgetID)
	if err := s.c.POST(url, &resModel, buf); err != nil {
		return nil, err
	}
	return resModel.Data.Account, nil
}
