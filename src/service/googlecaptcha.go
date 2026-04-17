package captcha

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func ValidateRecaptcha(secretKey string, g string) bool {

	url := "https://www.google.com/recaptcha/api/siteverify?secret=" + secretKey + "&response=" + g
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return false
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}

	var result GResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON")
		return false
	}

	return result.Success
}

// GenerateGoogleCaptchaHTML generates HTML for embedding Google reCAPTCHA in a web form.
func GenerateGoogleCaptchaHTML(siteKey string) string {
	// Example Google reCAPTCHA HTML code
	return `
		<script src="https://www.google.com/recaptcha/api.js" async defer></script>
		<div class="g-recaptcha" data-sitekey="` + siteKey + `"></div>
	`
}
