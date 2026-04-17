package controller

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"time"

	captcha "main.go/src/service"

	_ "main.go/src/service"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"main.go/database/repository"
)

const (
	dateFormat = "2006-01-02T15:04"
)

// generateToken creates a random base64 URL-encoded token
func generateToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// getEncryptionKey retrieves the AES encryption key from environment variables
func getEncryptionKey() ([]byte, error) {
	keyStr := os.Getenv("AES_ENCRYPTION_KEY")
	log.Printf("DEBUG: AES_ENCRYPTION_KEY length: %d", len(keyStr))
	if keyStr == "" {
		log.Printf("DEBUG: AES_ENCRYPTION_KEY environment variable not set")
		return nil, fmt.Errorf("AES_ENCRYPTION_KEY environment variable not set")
	}
	
	// Use the string directly as bytes (no base64 decoding)
	key := []byte(keyStr)
	
	// Ensure key is 32 bytes for AES-256
	log.Printf("DEBUG: Key length: %d bytes", len(key))
	if len(key) != 32 {
		log.Printf("DEBUG: Key length is %d, expected 32 bytes", len(key))
		return nil, fmt.Errorf("encryption key must be 32 bytes long for AES-256, got %d bytes", len(key))
	}
	
	log.Printf("DEBUG: Encryption key loaded successfully")
	return key, nil
}

// encryptText encrypts the given text using AES-GCM
func encryptText(plaintext string) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}
	
	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}
	
	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}
	
	// Encrypt the plaintext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptText decrypts the given encrypted text using AES-GCM
func decryptText(encryptedText string) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}
	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted text: %v", err)
	}
	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}
	// Extract nonce and encrypted data
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %v", err)
	}

	return string(plaintext), nil
}

func HomeGet(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "Hello, World!",
		"Text":  "",
	})
}

// HomeGet serves the main form HTML file
func invalidIndexSubmit(c *fiber.Ctx, msg string, text string) error {
	return c.Status(fiber.StatusBadRequest).Render("index", fiber.Map{
		"Title":    "Hello, World!",
		"ErrorMsg": msg,
		"Text":     text,
	})
}

func HomePost(c *fiber.Ctx) error {
	recaptcha := c.FormValue("g-recaptcha-response")
	fmt.Println(recaptcha)
	text := c.FormValue("text")

	if !captcha.ValidateRecaptcha(os.Getenv("RECAPTCHA_SECRET_KEY"), recaptcha) {
		return invalidIndexSubmit(c, "Invalid Recaptcha ", text)
	}

	// Get form data
	expirationDateStr := c.FormValue("expiration_date")
	password := c.FormValue("password") // Read password from form

	// Validate form data
	if text == "" || expirationDateStr == "" {
		return invalidIndexSubmit(c, "Text and expiration date are required", text)
	}

	// Validate password
	if !ValidatePassword(password) {
		return invalidIndexSubmit(c, "Password must be at least 8 characters long, contain an uppercase letter, a lowercase letter, a number, and a special character.", text)
	}

	// Encrypt the text before storing
	encryptedText, err := encryptText(text)
	if err != nil {
		log.Printf("Error encrypting text: %v", err)
		return invalidIndexSubmit(c, "Error processing secret text", text)
	}

	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return invalidIndexSubmit(c, "Error hashing password", text)
	}

	// Parse expiration date
	expirationDate := parseExpirationDate(expirationDateStr)

	// Validate that expiration date is in the future
	if expirationDate.Before(time.Now().UTC()) {
		return invalidIndexSubmit(c, "Expiration date must be in the future", text)
	}

	// Generate a unique token
	token, err := generateToken(32) // 32 bytes token
	if err != nil {
		return invalidIndexSubmit(c, "Error generating token", text)
	}

	// Create a new secret in the database with encrypted text
	secret := repository.CreateSecretParams{
		Token:          token,
		Text:           sql.NullString{String: encryptedText, Valid: true}, // Store encrypted text
		Password:       sql.NullString{String: hashedPassword, Valid: true}, 
		ExpirationDate: expirationDate,
	}

	// Create an instance of Queries using the global DB connection
	queries := repository.New(repository.DB)

	// Insert the secret
	_, err = queries.CreateSecret(c.Context(), secret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error inserting secret into database")
	}

	// Create a URL-safe link
	viewLink := fmt.Sprintf("http://%s/view?token=%s", os.Getenv("HOST_NAME"), token)

	// Log the generated link for debugging
	log.Printf("Generated view link: %s", viewLink)

	// Render the results page with the original (unencrypted) message for display
	return c.Render("results", fiber.Map{
		"Message": string(text), // Show original text to user
		"Link":    string(viewLink),
		"Token":   string(token),
	})
}

func BurnPost(c *fiber.Ctx) error {
	// Extract token from form data
	token := c.FormValue("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Token is required")
	}

	// Create an instance of Queries using the global DB connection
	queries := repository.New(repository.DB)

	// Mark the secret as burned in the database
	result, err := queries.MarkSecretAsBurned(c.Context(), token)
	if err != nil {
		log.Printf("Error burning secret: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error burning secret")
	}
	log.Printf("Secret burned: %v", result)

	// Redirect to the static HTML page
	return c.Render("burned", fiber.Map{})
}

func ViewGet(c *fiber.Ctx) error {
	// Extract token from query parameters
	token := c.Query("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Token is required")
	}

	// Create an instance of Queries using the global DB connection
	queries := repository.New(repository.DB)

	// Retrieve the secret from the database
	secret, err := queries.GetSecretByToken(c.Context(), token)
	if err != nil {
		log.Printf("Error retrieving secret: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving secret")
	}

	// Check if the secret has been burned
	if secret.IsBurned {
		return c.Status(fiber.StatusNotFound).SendString("Secret has been burned and is no longer available")
	}

	// Check if the secret has been viewed
	if secret.IsViewed {
		return c.Status(fiber.StatusForbidden).SendString("Secret has already been viewed")
	}

	// Check if the secret has expired
	if !secret.ExpirationDate.IsZero() && time.Now().After(secret.ExpirationDate) {
		// Mark the secret as burned if expired
		_, err = queries.MarkSecretAsBurned(c.Context(), token)
		if err != nil {
			log.Printf("Error marking secret as burned: %v", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Error marking secret as burned")
		}

		return c.Status(fiber.StatusNotFound).SendString("Secret has expired and is no longer available")
	}

	// Update the secret to mark it as viewed
	_, err = queries.MarkSecretAsViewed(c.Context(), token)
	if err != nil {
		log.Printf("Error marking secret as viewed: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error marking secret as viewed")
	}

	// Decrypt the text before displaying (if not password protected)
	var decryptedText string
	if secret.Text.Valid {
		decryptedText, err = decryptText(secret.Text.String)
		if err != nil {
			log.Printf("Error decrypting text: %v", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving secret content")
		}
	}

	// Render the template with the data
	return c.Render("viewsecret", fiber.Map{
		"Token":            token,
		"RequiresPassword": secret.Password.Valid,
		"Text":             decryptedText, // Show decrypted text
	})
}

// parseExpirationDate converts a string into a time.Time object
func parseExpirationDate(dateStr string) time.Time {
	date, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		log.Printf("Error parsing expiration date: %v", err)
		return time.Time{} // Return zero value if parsing fails
	}
	return date.UTC()
}

func VerifyPassword(c *fiber.Ctx) error {
	token := c.FormValue("token")
	password := c.FormValue("password")

	if token == "" || password == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Token and password are required")
	}

	queries := repository.New(repository.DB)

	secret, err := queries.GetSecretByToken(c.Context(), token)
	if err != nil {
		log.Printf("Error retrieving secret: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving secret")
	}

	// Verify the password
	if secret.Password.Valid {
		err = bcrypt.CompareHashAndPassword([]byte(secret.Password.String), []byte(password))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Incorrect password")
		}
	}

	// Decrypt the text before displaying
	var decryptedText string
	if secret.Text.Valid {
		decryptedText, err = decryptText(secret.Text.String)
		if err != nil {
			log.Printf("Error decrypting text: %v", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving secret content")
		}
	}

	// Password is correct, render the page with the decrypted secret text
	return c.Render("viewsecret", fiber.Map{
		"Text": decryptedText, // Show decrypted text
	})
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// validatePassword checks if the password meets specified criteria
func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false // Minimum length
	}
	if matched, _ := regexp.MatchString(`[A-Z]`, password); !matched {
		return false // At least one uppercase letter
	}
	if matched, _ := regexp.MatchString(`[a-z]`, password); !matched {
		return false // At least one lowercase letter
	}
	if matched, _ := regexp.MatchString(`[0-9]`, password); !matched {
		return false // At least one number
	}
	if matched, _ := regexp.MatchString(`[\W_]`, password); !matched {
		return false // At least one special character
	}
	return true
}