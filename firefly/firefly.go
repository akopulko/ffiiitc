package firefly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// struct for firefly http client
type FireFlyHttpClient struct {
	AppURL  string
	Timeout int
	Token   string
}

func NewFireFlyHttpClient(url, token string, timeout int) *FireFlyHttpClient {
	return &FireFlyHttpClient{
		AppURL:  url,
		Token:   token,
		Timeout: timeout,
	}
}

// set of structs for firefly transaction json data
type FireFlyTransaction struct {
	Description   string `json:"description"`
	Category      string `json:"category_name"`
	TransactionID string `json:"transaction_journal_id"`
}

type FireFlyTransactions struct {
	Transactions []FireFlyTransaction `json:"transactions"`
}

type FireFlyTransactionAttributes struct {
	Attributes FireFlyTransactions `json:"attributes"`
}

// set of structs for firefly pagination json
type FireFlyPaginationData struct {
	Total       int `json:"total"`
	Count       int `json:"count"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
}

type FireFlyPagination struct {
	Pagination FireFlyPaginationData `json:"pagination"`
}

// firefly transaction api response json
type FireFlyTransactionsResponse struct {
	Data []FireFlyTransactionAttributes `json:"data"`
	Meta FireFlyPagination              `json:"meta"`
}

// helper function to make http request to firefly api
// returns body
func (fc *FireFlyHttpClient) SendGetRequestWithToken(url, token string, timeout time.Duration) ([]byte, error) {
	client := http.Client{
		Timeout: timeout * time.Second, // Set a reasonable timeout for the request.
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Set the Authorization header with the Bearer token.
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

// helper function to make http request to firefly api
func (fc *FireFlyHttpClient) SendPutRequestWithToken(url, token string, data []byte, timeout time.Duration) ([]byte, error) {
	client := http.Client{
		Timeout: timeout * time.Second, // Set a reasonable timeout for the request.
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	// Set the Authorization header with the Bearer token.
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json") // Assuming JSON data, adjust the content type if needed.

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

func (fc *FireFlyHttpClient) UpdateTransactionCategory(id, category string) error {
	log.Printf("updating transaction: %s", id)

	trn := FireFlyTransactions{
		Transactions: []FireFlyTransaction{
			{
				TransactionID: id,
				Category:      category,
			},
		},
	}

	jsonData, err := json.Marshal(trn)
	if err != nil {
		log.Fatalf("Error marshaling JSON data: %v", err)
	}

	_, err = fc.SendPutRequestWithToken(
		fmt.Sprintf("%s/api/v1/transactions/%s", fc.AppURL, id),
		fc.Token,
		jsonData,
		time.Duration(fc.Timeout),
	)

	if err != nil {
		return err
	}

	return nil
}

// get all transactions
// returns slice of strings "transaction description, category"
func (fc *FireFlyHttpClient) GetTransactions() ([]string, error) {
	var pageIndex int
	log.Println("get first page of transactions")
	res, err := fc.SendGetRequestWithToken(
		fmt.Sprintf("%s/api/v1/transactions?page=1", fc.AppURL),
		fc.Token,
		time.Duration(fc.Timeout),
	)
	if err != nil {
		return nil, err
	}

	var data FireFlyTransactionsResponse
	err = json.Unmarshal(res, &data)
	if err != nil {
		return nil, err
	}
	var transactions []string
	for _, value := range data.Data {
		for _, trnval := range value.Attributes.Transactions {
			trn := fmt.Sprintf("%s,%s", trnval.Category, trnval.Description)
			transactions = append(transactions, trn)
		}
	}

	log.Printf("transactions total pages: %d", data.Meta.Pagination.TotalPages)

	if data.Meta.Pagination.TotalPages > 1 {
		log.Println("transactions more than 1 page available. iterating")
		pageIndex = 2
		for pageIndex <= data.Meta.Pagination.TotalPages {
			res, err := fc.SendGetRequestWithToken(
				fmt.Sprintf("%s/api/v1/transactions?page=%d", fc.AppURL, pageIndex),
				fc.Token,
				time.Duration(fc.Timeout),
			)
			if err != nil {
				return nil, err
			}
			var data FireFlyTransactionsResponse
			err = json.Unmarshal(res, &data)
			if err != nil {
				return nil, err
			}
			for _, value := range data.Data {
				for _, trnval := range value.Attributes.Transactions {
					trn := fmt.Sprintf("%s,%s", trnval.Category, trnval.Description)
					transactions = append(transactions, trn)
				}
			}
			log.Printf("page %d...", pageIndex)
			pageIndex++
		}
	}
	return transactions, err
}
