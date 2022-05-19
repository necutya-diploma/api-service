package fondy

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/structs"
)

const (
	contentTypeApplicationJson = "application/json"
	checkoutUrl                = "https://pay.fondy.eu/api/checkout/url/"

	signatureDelimiter = "|"
	successRespStatus  = "success"
	failureRespStatus  = "failure"

	USDCurrency = "USD"
	MonthPeriod = "month"

	yesState = "Y"

	JQLDateFormat = "2006-02-01"
)

var (
	ErrInvalidResponseStatus = errors.New("invalid response status")
)

type apiCheckoutRequest struct {
	Request *checkoutRequest `json:"request"`
}

type checkoutRequest struct {
	OrderID           string `json:"order_id"`
	MerchantID        int64  `json:"merchant_id"`
	OrderDesc         string `json:"order_desc"`
	Signature         string `json:"signature"`
	Amount            int64  `json:"amount"`
	Currency          string `json:"currency"`
	SenderEmail       string `json:"sender_email"`
	Subscription      string `json:"subscription,omitempty"`
	ProductID         string `json:"product_id"`
	ServerCallbackURL string `json:"server_callback_url"`
	ResponseURL       string `json:"response_url"`
}

func (cr *checkoutRequest) setSignature(password string) {
	params := structs.Map(cr)
	cr.Signature = generateSignature(params, password)
}

type apiCheckoutResponse struct {
	Response *checkoutResponse `json:"response"`
}

type checkoutResponse struct {
	ResponseStatus string `json:"response_status"`
	CheckoutURL    string `json:"checkout_url"`
	SenderEmail    string `json:"sender_email"`
	ErrorMessage   string `json:"error_message"`
	ErrorCode      int64  `json:"error_code"`
}

type Callback struct {
	OrderId                 string      `json:"order_id"`
	MerchantId              int         `json:"merchant_id"`
	Amount                  string      `json:"amount"`
	Currency                string      `json:"currency"`
	OrderStatus             string      `json:"order_status"`    // created; processing; declined; approved; expired; reversed;
	ResponseStatus          string      `json:"response_status"` // 1) success; 2) failure
	Signature               string      `json:"signature"`
	TranType                string      `json:"tran_type"`
	SenderCellPhone         string      `json:"sender_cell_phone"`
	SenderAccount           string      `json:"sender_account"`
	CardBin                 int         `json:"card_bin"`
	MaskedCard              string      `json:"masked_card"`
	CardType                string      `json:"card_type"`
	RRN                     string      `json:"rrn"`
	ApprovalCode            string      `json:"approval_code"`
	ResponseCode            interface{} `json:"response_code"`
	ResponseDescription     string      `json:"response_description"`
	ReversalAmount          string      `json:"reversal_amount"`
	SettlementAmount        string      `json:"settlement_amount"`
	SettlementCurrency      string      `json:"settlement_currency"`
	OrderTime               string      `json:"order_time"`
	SettlementDate          string      `json:"settlement_date"`
	ECI                     string      `json:"eci"`
	Fee                     string      `json:"fee"`
	PaymentSystem           string      `json:"payment_system"`
	SenderEmail             string      `json:"sender_email"`
	PaymentId               int         `json:"payment_id"`
	ActualAmount            string      `json:"actual_amount"`
	ActualCurrency          string      `json:"actual_currency"`
	MerchantData            string      `json:"merchant_data"`
	VerificationStatus      string      `json:"verification_status"`
	Rectoken                string      `json:"rectoken"`
	RectokenLifetime        string      `json:"rectoken_lifetime"`
	ProductId               string      `json:"product_id"`
	AdditionalInfo          string      `json:"additional_info"`
	ResponseSignatureString string      `json:"response_signature_string"`
}

func (c Callback) Success() bool {
	return c.ResponseStatus == "success"
}

func (c Callback) PaymentApproved() bool {
	return c.OrderStatus == "approved"
}

func generateSignature(params map[string]interface{}, password string) string {
	keys := make([]string, 0, len(params))

	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	values := make([]string, 0, len(params))

	for _, k := range keys {
		if value, ok := params[k].(int64); ok {
			params[k] = strconv.FormatInt(value, 10)
		}

		if value, ok := params[k].(string); ok && value != "" {
			values = append(values, value)
		}
	}

	values = append([]string{password}, values...)

	return hash(strings.Join(values, signatureDelimiter))
}

func hash(value string) string {
	h := sha1.New()

	h.Write([]byte(value))

	return fmt.Sprintf("%x", h.Sum(nil))
}
