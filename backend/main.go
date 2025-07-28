package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Make sure to replace this with your actual spreadsheet ID

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	})

	// Endpoint to add an email to the Google Sheet
	r.POST("/add-email", func(c *gin.Context) {

		var json struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := addEmailToSheet(json.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email added successfully"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}
	r.Run(":" + port)
}

func addEmailToSheet(email string) error {
	ctx := context.Background()
	spreadsheetId := os.Getenv("spreadsheetId")
	// Read the service account credentials from the environment variable
	creds := os.Getenv("GOOGLE_CREDENTIALS")
	if creds == "" {
		return fmt.Errorf("GOOGLE_CREDENTIALS environment variable not set")
	}
	b := []byte(creds)

	// Configure the JWT with the correct scope
	config, err := google.JWTConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	// Create a new Google Sheets service client
	client := config.Client(ctx)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	// The A1 notation of the values to check.
	range2 := "Sheet1" // Change to your sheet name if different

	// Get the existing values in the sheet
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, range2).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	// Check if the email already exists in the sheet
	for _, row := range resp.Values {
		if len(row) > 0 && row[0] == email {
			return fmt.Errorf("email already exists in the sheet")
		}
	}

	// The data to append.
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, []interface{}{email})

	// Append the new row.
	_, err = srv.Spreadsheets.Values.Append(spreadsheetId, range2, &vr).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to append data to sheet: %v", err)
	}

	return nil
}
