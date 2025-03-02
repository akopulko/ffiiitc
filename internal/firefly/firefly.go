package firefly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-pkgz/lgr"
)

const (
	fireflyAPIPrefix = "api/v1"
)

type Timeout time.Duration

// struct for firefly http client
type FireFlyHttpClient struct {
	AppURL string
	//Timeout Timeout
	Timeout Timeout
	Token   string
	logger  *lgr.Logger
}

// set of structs for firefly transaction json data
type FireFlyTransaction struct {
	Description   string `json:"description"`
	Category      string `json:"category_name"`
	TransactionID string `json:"transaction_journal_id"`
}

type FireFlyTransactions struct {
	FireWebHooks bool                 `json:"fire_webhooks"`
	Id           string                  `json:"id"`
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

func NewFireFlyHttpClient(url, token string, timeout Timeout, l *lgr.Logger) *FireFlyHttpClient {
	return &FireFlyHttpClient{
		AppURL:  url,
		Token:   token,
		Timeout: timeout,
		logger:  l,
	}
}

// helper function to make http request to firefly api
// returns body
func (fc *FireFlyHttpClient) sendRequestWithToken(method, url, token string, data []byte) ([]byte, error) {
	client := http.Client{
		Timeout: time.Duration(fc.Timeout) * time.Second, // Set a reasonable timeout for the request.
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	// Set the Authorization header with the Bearer token.
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/vnd.api+json")

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

// SendGetRequestWithToken sends an HTTP GET request to the FireFly API with a token.
func (fc *FireFlyHttpClient) SendGetRequestWithToken(url, token string) ([]byte, error) {
	return fc.sendRequestWithToken(http.MethodGet, url, token, nil)
}

// SendPutRequestWithToken sends an HTTP PUT request to the FireFly API with a token and data.
func (fc *FireFlyHttpClient) SendPutRequestWithToken(url, token string, data []byte) ([]byte, error) {
	return fc.sendRequestWithToken(http.MethodPut, url, token, data)
}

func (fc *FireFlyHttpClient) UpdateTransactionCategory(id, trans_id, category string) error {
	//log.Printf("updating transaction: %s", id)

	trn := FireFlyTransactions{
		FireWebHooks: false,
		Id: id,
		Transactions: []FireFlyTransaction{
			{
				TransactionID: trans_id,
				Category:      category,
			},
		},
	}

	//log.Printf("trn data: %v", trn)
	fc.logger.Logf("DEBUG trn data: %v", trn)

	jsonData, err := json.Marshal(trn)
	if err != nil {
		//log.Fatalf("Error marshaling JSON data: %v", err)
		return err
	}

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, jsonData, "", "    ")
	//log.Printf("json sent: %s", prettyJSON.String())
	fc.logger.Logf("DEBUG json sent: %s", prettyJSON.String())

	res, err := fc.SendPutRequestWithToken(
		fmt.Sprintf("%s/%s/transactions/%s", fc.AppURL, fireflyAPIPrefix, id),
		fc.Token,
		jsonData,
	)

	if err != nil {
		return err
	}

	//debug
	prettyJSON = bytes.Buffer{}
	json.Indent(&prettyJSON, res, "", "    ")
	//log.Println(prettyJSON.String())
	fc.logger.Logf("DEBUG json received: %s", prettyJSON.String())

	return nil
}

func buildCategoryDescriptionSlice(data FireFlyTransactionsResponse) []string {
	var res []string
	for _, value := range data.Data {
		for _, trnval := range value.Attributes.Transactions {
			trn := fmt.Sprintf("%s,%s", trnval.Category, trnval.Description)
			res = append(res, trn)
		}
	}
	return res
}

func buildTransactionsDataset(data FireFlyTransactionsResponse) [][]string {
	var res [][]string
	for _, value := range data.Data {
		for _, trnval := range value.Attributes.Transactions {
			//trn := fmt.Sprintf("%s,%s", trnval.Category, trnval.Description)
			trn := []string{trnval.Category, trnval.Description}
			res = append(res, trn)
		}
	}
	return res
}

// get all transactions
// returns slice of strings "transaction description, category"
func (fc *FireFlyHttpClient) GetTransactions() ([]string, error) {
	var pageIndex int
	//log.Println("get first page of transactions")
	fc.logger.Logf("INFO get first page of transactions")
	res, err := fc.SendGetRequestWithToken(
		fmt.Sprintf("%s/api/v1/transactions?page=1", fc.AppURL),
		fc.Token,
	)
	if err != nil {
		return nil, err
	}

	var data FireFlyTransactionsResponse
	err = json.Unmarshal(res, &data)
	if err != nil {
		return nil, err
	}
	resSlice := buildCategoryDescriptionSlice(data)

	//log.Printf("transactions total pages: %d", data.Meta.Pagination.TotalPages)
	fc.logger.Logf("INFO transactions total pages: %d", data.Meta.Pagination.TotalPages)
	if data.Meta.Pagination.TotalPages > 1 {
		//log.Println("transactions more than 1 page available. iterating")
		fc.logger.Logf("INFO transactions more than 1 page available. iterating")
		pageIndex = 2
		for pageIndex <= data.Meta.Pagination.TotalPages {
			res, err := fc.SendGetRequestWithToken(
				fmt.Sprintf("%s/%s/transactions?page=%d", fc.AppURL, fireflyAPIPrefix, pageIndex),
				fc.Token,
			)
			if err != nil {
				return nil, err
			}
			var data FireFlyTransactionsResponse
			err = json.Unmarshal(res, &data)
			if err != nil {
				return nil, err
			}
			resSlice = append(resSlice, buildCategoryDescriptionSlice(data)...)
			//log.Printf("page %d...", pageIndex)
			fc.logger.Logf("INFO page %d...", pageIndex)
			pageIndex++
		}
	}
	// return transactions, err
	return resSlice, err
}

func (fc *FireFlyHttpClient) GetTransactionsDataset() ([][]string, error) {
	var pageIndex int
	fc.logger.Logf("INFO get first page of transactions")
	res, err := fc.SendGetRequestWithToken(
		fmt.Sprintf("%s/api/v1/transactions?page=1", fc.AppURL),
		fc.Token,
	)
	if err != nil {
		return nil, err
	}

	var data FireFlyTransactionsResponse
	err = json.Unmarshal(res, &data)
	if err != nil {
		return nil, err
	}
	resSlice := buildTransactionsDataset(data)

	fc.logger.Logf("INFO transactions total pages: %d", data.Meta.Pagination.TotalPages)
	if data.Meta.Pagination.TotalPages > 1 {
		fc.logger.Logf("INFO transactions more than 1 page available. iterating")
		pageIndex = 2
		for pageIndex <= data.Meta.Pagination.TotalPages {
			res, err := fc.SendGetRequestWithToken(
				fmt.Sprintf("%s/%s/transactions?page=%d", fc.AppURL, fireflyAPIPrefix, pageIndex),
				fc.Token,
			)
			if err != nil {
				return nil, err
			}
			var data FireFlyTransactionsResponse
			err = json.Unmarshal(res, &data)
			if err != nil {
				return nil, err
			}
			resSlice = append(resSlice, buildTransactionsDataset(data)...)
			fc.logger.Logf("INFO page %d...", pageIndex)
			pageIndex++
		}
	}
	return resSlice, err
}
