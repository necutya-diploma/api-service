package fondy

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type PaymentsManager struct {
	merchantID       int64
	merchantPassword string
	currency         string
	language         string

	responseURL string
	callbackUrl string
}

func NewPaymentManager(
	merchantID int64,
	merchantPassword, currency, language string,
	responseURL, callbackURL string,
) *PaymentsManager {
	return &PaymentsManager{
		merchantID:       merchantID,
		merchantPassword: merchantPassword,
		currency:         currency,
		language:         language,
		responseURL:      responseURL,
		callbackUrl:      callbackURL,
	}
}

func (pm *PaymentsManager) GenerateSubscriptionCheckout(
	orderID, orderDesc, senderEmail, productID string,
	amount int64,
) (string, error) {
	checkoutReq := checkoutRequest{
		OrderID:           orderID,
		MerchantID:        pm.merchantID,
		OrderDesc:         orderDesc,
		ProductID:         productID,
		Amount:            amount,
		Currency:          pm.currency,
		SenderEmail:       senderEmail,
		ServerCallbackURL: pm.callbackUrl,
		ResponseURL:       pm.responseURL,
	}

	checkoutReq.setSignature(pm.merchantPassword)

	apiRequest := apiCheckoutRequest{
		Request: &checkoutReq,
	}

	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		checkoutUrl,
		contentTypeApplicationJson,
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	apiResponse := apiCheckoutResponse{}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return "", err
	}

	switch apiResponse.Response.ResponseStatus {
	case successRespStatus:
		return apiResponse.Response.CheckoutURL, nil
	case failureRespStatus:
		return "", errors.New(apiResponse.Response.ErrorMessage)
	}

	return "", ErrInvalidResponseStatus
}

func (pm *PaymentsManager) ValidateCallback(input interface{}) error {
	_, ok := input.(Callback)
	if !ok {
		return errors.New("invalid callback data")
	}

	return nil
}
