package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"bytes"
	"context"

	"github.com/Rohnson95/weathermonitoring/backend/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)



var jwtKey []byte

// Struct for SensorRequests
type SensorRequest struct {
    Name string `json:"name"`
    Latitude float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Description string `json:"description"`
}
// Struct for WebhookRequests
type Claims struct {
    UserID uint `json:"user_id"`
    jwt.RegisteredClaims
}
type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Company string `json:"company"`
}
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type TokenResponse struct {
    Token string `json:"token"`
}
type WeatherResponse struct {
    Weather     string  `json:"weather"`
    Temperature float64 `json:"temperature"`
}

type WeatherForecast struct {
    TimeSeries []struct {
        ValidTime  string `json:"validTime"`
        Parameters []struct {
            Name  string  `json:"name"`
            Value float64 `json:"value"`
            } `json:"parameters"`
            } `json:"timeSeries"`
        }
        
        type Warning struct {
            ID                int            `json:"id"`
            Event             Event          `json:"event"`
            NormalProbability bool           `json:"normalProbability"`
            AreaName          AreaName       `json:"areaName,omitempty"`
            WarningAreas      []WarningArea  `json:"warningAreas"`
            Descriptions      []Description  `json:"descriptions,omitempty"`
        }
        
        type Event struct {
            En                string           `json:"en"`
            Sv                string           `json:"sv"`
            Code              string           `json:"code"`
            MhoClassification MhoClassification `json:"mhoClassification"`
        }
        
        type MhoClassification struct {
            En   string `json:"en"`
            Sv   string `json:"sv"`
            Code string `json:"code"`
        }
        
        type AreaName struct {
            En string `json:"en"`
            Sv string `json:"sv"`
        }
        
type WarningArea struct {
    ID                 int            `json:"id"`
    Area               GeoJSONFeature        `json:"area"`
    AreaName           AreaName       `json:"areaName,omitempty"`
    ApproximateStart   string         `json:"approximateStart"`
    ApproximateEnd     string         `json:"approximateEnd,omitempty"`
    NormalProbability  bool           `json:"normalProbability"`
    WarningLevel       WarningLevel   `json:"warningLevel"`
    EventDescription   EventDescription `json:"eventDescription"`
    Published          string         `json:"published"`
    Descriptions       []Description  `json:"descriptions,omitempty"`
    AffectedAreas      []AffectedArea `json:"affectedAreas"`
}

    type GeoJSONFeature struct {
        Type     string `json:"type"` 
        Geometry struct {
            Type        string        `json:"type"` 
            Coordinates interface{}   `json:"coordinates"`
            } `json:"geometry"`
            Properties map[string]interface{} `json:"properties"`
        }
        
        type WarningLevel struct {
            En   string `json:"en"`
            Sv   string `json:"sv"`
            Code string `json:"code"`
        }
        
        type EventDescription struct {
            En   string `json:"en"`
            Sv   string `json:"sv"`
            Code string `json:"code"`
        }
        
        type Description struct {
            Title Title `json:"title"`
            Text  Text  `json:"text"`
        }
        
        type Title struct {
            En   string `json:"en"`
            Sv   string `json:"sv"`
            Code string `json:"code"`
}

type Text struct {
    En   string `json:"en"`
    Sv   string `json:"sv"`
}

type AffectedArea struct {
    ID int    `json:"id"`
    Sv string `json:"sv"`
    En string `json:"en"`
}

var(
    db *gorm.DB
    cachedWarnings []Warning
    warningsMutex sync.Mutex
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
        return
    }

    // Hash the password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Error hashing password", http.StatusInternalServerError)
        return
    }

    // Check if company exists
    var company models.Company
    result := db.Where("name = ?", req.Company).First(&company)
    if errors.Is(result.Error, gorm.ErrRecordNotFound){
        //Create New Company
        company = models.Company{Name: req.Company}
        if err := db.Create(&company).Error; err != nil {
            http.Error(w, "Error creating company", http.StatusInternalServerError)
            return
        }
    } else if result.Error != nil {
        http.Error(w, "Database Error", http.StatusInternalServerError)
        return
    }

    // Create User
    user := models.User{
        Email: req.Email,
        PasswordHash: string(hashedPassword),
        CompanyID: company.ID,
    }
    if err := db.Create(&user).Error; err != nil {
        http.Error(w, "Error creating user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}
// LoginHandler handles user login and JWT issuance
func LoginHandler(w http.ResponseWriter, r *http.Request){
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
        return
    }

    // Find user by email
    var user models.User
    if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    // Compare password
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    // Create JWT token
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        UserID: user.ID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt: jwt.NewNumericDate(time.Now()),
            Issuer: "weather-iot",
        },
    }

    // Create token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        http.Error(w, "Error generating token", http.StatusInternalServerError)
        return
    }

    // Send Token
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(TokenResponse{Token: tokenString})
}

// Middleware to authenticate requests
func Authenticate(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request){
        // Get token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
            return
        }

        //Expecting header format: Bearer <token>
        var tokenString string
        fmt.Sscanf(authHeader, "Bearer %s", &tokenString)

        claims := &Claims{}

        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error){
            return jwtKey, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid Token", http.StatusUnauthorized)
            return
        }

        // Attach user ID to the request context
        ctx := r.Context()
        ctx = contextWithUserID(ctx, claims.UserID)
        r = r.WithContext(ctx)

        next(w,r)
    }
}
type contextKey string

const userIDKey contextKey = "userID"

func contextWithUserID(ctx context.Context, userID uint) context.Context {
    return context.WithValue(ctx, userIDKey, userID)
}

func getUserIDFromContext(ctx context.Context) (uint, error){
    userID, ok := ctx.Value(userIDKey).(uint)
    if !ok {
        return 0, errors.New("user ID not found in context")
    }
    return userID, nil
}

func SensorHandler(w http.ResponseWriter, r *http.Request){
    userID, err := getUserIDFromContext(r.Context())
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    switch r.Method {
    case http.MethodGet:
        handleGetSensors(w, r, userID)
    case http.MethodPost:
        handleCreateSensor(w, r, userID)
    case http.MethodPut:
        handleUpdateSensor(w, r, userID)
    case http.MethodDelete:
        handleDeleteSensor(w, r, userID)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func handleGetSensors(w http.ResponseWriter, r *http.Request, userID uint) {
    // Get user's company
    var user models.User
    if err := db.Preload("Company").First(&user, userID).Error; err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    // Fetch sensors belonging to the company
    var sensors []models.Sensor
    if err := db.Where("company_id = ?", user.CompanyID).Find(&sensors).Error; err != nil {
        http.Error(w, "Error fetching sensors", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK) // Changed from http.StatusCreated to http.StatusOK
    json.NewEncoder(w).Encode(sensors)
}

func handleCreateSensor(w http.ResponseWriter, r *http.Request, userID uint){
    var req SensorRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
        return
    }

    // Get user's company
    var user models.User
    if err := db.First(&user, userID).Error; err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    sensor := models.Sensor{
        Name: req.Name,
        Latitude: req.Latitude,
        Longitude: req.Longitude,
        Description: req.Description,
        Status: "OK", // Default status
        CompanyID: user.CompanyID,
    }

    if err := db.Create(&sensor).Error; err != nil {
        http.Error(w, "Error creating sensor", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(sensor)
}


func handleUpdateSensor(w http.ResponseWriter, r *http.Request, userID uint){
    // Assume sensor ID is passed as a query parameter

    sensorID := r.URL.Query().Get("id")
    if sensorID == "" {
        http.Error(w, "Missing sensor ID", http.StatusBadRequest)
        return
    }

    var req SensorRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
        return
    }

    var user models.User
    if err := db.First(&user, userID).Error; err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    // Find the sensor
    var sensor models.Sensor
    if err := db.Where("id = ? AND company_id = ?", sensorID, user.CompanyID).First(&sensor).Error; err != nil {
        http.Error(w, "Sensor not found", http.StatusNotFound)
        return
    }

    // Update Sensorfields
    sensor.Name = req.Name
    sensor.Latitude = req.Latitude
    sensor.Longitude = req.Longitude
    sensor.Description = req.Description

    if err := db.Save(&sensor).Error; err != nil {
        http.Error(w, "Error updating sensor", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(sensor)
}

func handleDeleteSensor(w http.ResponseWriter, r *http.Request, userID uint){
    sensorID := r.URL.Query().Get("id")
    if sensorID == "" {
        http.Error(w, "Missing sensor ID", http.StatusBadRequest)
        return
    }
    // Users company
    var user models.User
    if err := db.First(&user, userID).Error; err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }

    if err := db.Where("id = ? AND company_id = ?", sensorID, user.CompanyID).Delete(&models.Sensor{}).Error; err != nil {
        http.Error(w, "Error deleting Sensor", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

// WebhookHandler

type WebhookRequest struct {
    URL string `json:"url"`
}

func WebhookHandler(w http.ResponseWriter, r *http.Request){
    userID, err := getUserIDFromContext(r.Context())
    if err != nil{
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    switch r.Method{
    case http.MethodGet:
        handleGetWebhooks(w, r, userID)
    case http.MethodPost:
        handleCreateWebhook(w, r, userID)
    case http.MethodDelete:
        handleDeleteWebhook(w, r, userID)

    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func handleGetWebhooks(w http.ResponseWriter, r *http.Request, userID uint){
    var webhooks []models.Webhook
    if err := db.Where("user_id = ?", userID).Find(&webhooks).Error; err != nil {
        http.Error(w, "Error fetching webhooks", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(webhooks)
}

func handleCreateWebhook(w http.ResponseWriter, r *http.Request, userID uint){
    var req WebhookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
        return
    }

    webhook := models.Webhook{
        URL: req.URL,
        UserID: userID,
    }

    if err := db.Create(&webhook).Error; err != nil {
        http.Error(w, "Error creating Webhook", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(webhook)
}

func handleDeleteWebhook(w http.ResponseWriter, r *http.Request, userID uint){
    webhookID := r.URL.Query().Get("id")
    if webhookID == "" {
        http.Error(w, "Missing webhook ID", http.StatusBadRequest)
        return
    }

    if err := db.Where("id = ? AND user_id = ?", webhookID, userID).Delete(&models.Webhook{}).Error; err != nil {        
        http.Error(w, "Error deleting Webhook", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

// Webhook notification logic
func periodicWeatherUpdate(){
    ticker := time.NewTicker(15 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            warnings, err := fetchWeatherWarnings()
            if err != nil {
                log.Printf("Error fetching weather warnings: %v", err)
                continue
            }

            warningsMutex.Lock()
            cachedWarnings = warnings
            warningsMutex.Unlock()

            log.Printf("Fetched and cached %d warnings", len(warnings))

            // process warnings and send webhook notifications
            go processWarningsAndNotify(warnings)
        }
    }
}

func processWarningsAndNotify(warnings []Warning) {
    // Iterate through warnings to find orange or red
    for _, warning := range warnings {
        level := strings.ToUpper(warning.Event.Code) // Assuming code represents the level
        if level != "ORANGE" && level != "RED" {
            continue
        }

        // For each warning area, find sensors within the area
        for _, warningArea := range warning.WarningAreas {
            // Fetch sensors in this area
            // Implement spatial queries as needed

            // Fetch all companies (assuming multiple companies can have sensors)
            var companies []models.Company
            if err := db.Preload("Sensors").Find(&companies).Error; err != nil {
                log.Printf("Error fetching companies: %v", err)
                continue
            }

            for _, company := range companies {
                for _, sensor := range company.Sensors {
                    // Check if sensor is within the warning area
                    // Implement point-in-polygon logic or use spatial extensions like PostGIS
                    // For simplicity, let's assume we have a function isSensorInArea
                    if isSensorInArea(sensor.Latitude, sensor.Longitude, warningArea.Area) {
                        // Update sensor status
                        sensor.Status = level
                        db.Save(&sensor)

                        // Find user webhooks associated with the company
                        var webhooks []models.Webhook
                        // Fetch all users in the company
                        var users []models.User
                        if err := db.Where("company_id = ?", company.ID).Find(&users).Error; err != nil {
                            log.Printf("Error fetching users for company %d: %v", company.ID, err)
                            continue
                        }

                        // Collect all webhooks for these users
                        for _, user := range users {
                            var userWebhooks []models.Webhook
                            if err := db.Where("user_id = ?", user.ID).Find(&userWebhooks).Error; err != nil {
                                log.Printf("Error fetching webhooks for user %d: %v", user.ID, err)
                                continue
                            }
                            webhooks = append(webhooks, userWebhooks...)
                        }

                        // Send notifications
                        for _, webhook := range webhooks {
                            go sendWebhookNotification(webhook.URL, sensor, warning)
                        }
                    }
                }
            }
        }
    }
}


func isSensorInArea(lat, long float64, area GeoJSONFeature) bool {
    return false // placeholder
    }

func sendWebhookNotification(url string, sensor models.Sensor, warning Warning) {
    payload := map[string]interface{}{
		"sensor_id":   sensor.ID,
		"sensor_name": sensor.Name,
		"status":      sensor.Status,
		"warning":     warning.Event.En,
		"timestamp":   time.Now(),
	}

    data, err := json.Marshal(payload)
    if err != nil{
        log.Printf("Error marshalling payload: %v", err)
        return
    }

    resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
    if err != nil {
        log.Printf("Error sending webhook notification: %v", err)
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 300 {
        log.Printf("webhook %s responded with status %d", url, resp.StatusCode)
    }
}

func fetchWeatherWarnings() ([]Warning, error) {
    url := "https://opendata-download-warnings.smhi.se/ibww/api/version/1/warning.json"
    resp, err := http.Get(url)
    if err != nil {
        log.Printf("Error fetching warnings from SMHI: %v", err)
        return nil, fmt.Errorf("failed to fetch warnings: %w", err)
    }
    defer resp.Body.Close()

    var warnings []Warning
    if err := json.NewDecoder(resp.Body).Decode(&warnings); err != nil {
        log.Printf("Error decoding warnings response: %v", err)
        return nil, fmt.Errorf("failed to decode warnings: %w", err)
    }

    log.Printf("Fetched %d warnings from SMHI", len(warnings))
    for i, warning := range warnings {
        log.Printf("Warning %d: %+v", i, warning)
    }
    return warnings, nil
}

func filterAndPaginateWarnings(warnings []Warning, eventType, areaName string, page, pageSize int) []Warning {
    filteredWarnings := []Warning{}
    for _, warning := range warnings {
        // Filter by event type (optional)
        if eventType != "" && warning.Event.Code != eventType {
            continue
        }

        // Filter by area name
        if areaName != "" {
            areaFound := false
            // Create new slice for matching warning areas
            matchingWarningAreas := []WarningArea{}

            for _, warningArea := range warning.WarningAreas {
                areaMatches := false
                matchingAffectedAreas := []AffectedArea{}

                for _, affectedArea := range warningArea.AffectedAreas {
                    if strings.Contains(strings.ToLower(affectedArea.Sv), strings.ToLower(areaName)) {
                        areaFound = true
                        areaMatches = true
                        matchingAffectedAreas = append(matchingAffectedAreas, affectedArea)
                    }
                }

                if areaMatches {
                    // Only include matching affected areas
                    warningArea.AffectedAreas = matchingAffectedAreas
                    matchingWarningAreas = append(matchingWarningAreas, warningArea)
                }
            }

            if areaFound {
                // Only include matching warning areas
                warning.WarningAreas = matchingWarningAreas
                filteredWarnings = append(filteredWarnings, warning)
            }
            continue
        }

        filteredWarnings = append(filteredWarnings, warning)
    }

    // Paginate warnings
    start := (page - 1) * pageSize
    if start >= len(filteredWarnings) {
        return []Warning{}
    }

    end := start + pageSize
    if end > len(filteredWarnings) {
        end = len(filteredWarnings)
    }

    return filteredWarnings[start:end]
}

func weatherWarningHandler(w http.ResponseWriter, r *http.Request) {
    warningsMutex.Lock()
    warnings := cachedWarnings
    warningsMutex.Unlock()

    log.Printf("Serving %d cached warnings", len(warnings))

    eventType := r.URL.Query().Get("eventType")
    areaName := r.URL.Query().Get("areaName")
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 10
    }

    filteredWarnings := filterAndPaginateWarnings(warnings, eventType, areaName, page, pageSize)
    response := struct {
        Warnings []Warning `json:"warnings"`
    }{
        Warnings: filteredWarnings,
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        log.Printf("Error encoding response: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}


func main() {
   // Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Proceeding with environment variables.")
	}

	// Retrieve environment variables
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	smhiAPIURL := os.Getenv("SMHI_API_URL")
	serverPort := os.Getenv("SERVER_PORT")
	frontendURL := os.Getenv("FRONTEND_URL")

	// Validate essential environment variables
	if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" || dbPort == "" || jwtSecret == "" || smhiAPIURL == "" || serverPort == "" || frontendURL == "" {
		log.Fatal("One or more required environment variables are missing.")
	}

	// Initialize JWT key
	jwtKey = []byte(jwtSecret)

	// Create Data Source Name (DSN) for PostgreSQL
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName, dbPort)

	// Connect to the database
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to DataBase: %v", err)
	}

	// Migrate all models
	err = db.AutoMigrate(&models.User{}, &models.Company{}, &models.Webhook{}, &models.Sensor{})
	if err != nil {
		log.Fatalf("Failed to migrate DataBase: %v", err)
	}

	mux := http.NewServeMux()

	// Initial fetch of weather warnings
	warnings, err := fetchWeatherWarnings()
	if err != nil {
		log.Fatalf("Error fetching initial weather warnings: %v", err)
	}

	warningsMutex.Lock()
	cachedWarnings = warnings
	warningsMutex.Unlock()

	log.Printf("Initial fetch: cached %d warnings", len(warnings))

	// Start periodic updates
	go periodicWeatherUpdate()

	// Register all your routes without CORS (handled globally)
	mux.HandleFunc("/api/register", RegisterHandler)
	mux.HandleFunc("/api/login", LoginHandler)
	mux.HandleFunc("/api/weather-warnings", weatherWarningHandler)
	mux.HandleFunc("/api/sensors", Authenticate(SensorHandler))
	mux.HandleFunc("/api/webhooks", Authenticate(WebhookHandler))

	// Configure CORS with rs/cors
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{frontendURL}, // Your frontend URL from .env
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	fmt.Printf("Starting server on :%s\n", serverPort)
	if err := http.ListenAndServe(":"+serverPort, handler); err != nil {
		log.Fatal(err)
    }
}