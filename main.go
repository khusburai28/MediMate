package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
	"os"
	"bytes"
	"encoding/json"
	"mime/multipart"
	"encoding/base64"
	"io"
	"fmt"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Username string             `bson:"username"`
	Password string             `bson:"password"` // Stored as SHA256 hash
	Role     string             `bson:"role"`     // Only "patient" now
}

type Prescription struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	PatientID   string            `bson:"patient_id"`
	ImagePath   string            `bson:"image_path"`
	Analysis    string            `bson:"analysis"`
	UploadDate  time.Time         `bson:"upload_date"`
}

type Medicine struct {
	Name     string `json:"name"`
	Dosage   string `json:"dosage"`
	Disease  string `json:"disease"`
	Usage    string `json:"usage"`
}

type PageData struct {
	User         string
	Role         string
	Results      []Medicine
	Prescriptions []Prescription
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

type ChatRequest struct {
	Message string `json:"message"`
}

type DiseasePredictionRequest struct {
	Age           string `json:"age"`
	Gender        string `json:"gender"`
	Symptoms      string `json:"symptoms"`
	MedicalHistory string `json:"medical_history"`
}

var (
	mutex          sync.Mutex
	client         *mongo.Client
	usersColl      *mongo.Collection
	prescriptionsColl *mongo.Collection
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func getMongoURI() string {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}
	
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Println("MONGODB_URI not found in .env, using default")
		return "mongodb://localhost:27017"
	}
	
	return mongoURI
}

func askGemini(prompt string, file ...multipart.File) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	url := os.Getenv("GEMINI_API_URL") + "?key=" + apiKey

	var requestBody map[string]interface{}

	if len(file) > 0 && file[0] != nil {
		uploadedFile := file[0]
		defer uploadedFile.Close()

		fileData, err := io.ReadAll(uploadedFile)
		if err != nil {
			return "", fmt.Errorf("error reading uploaded file: %w", err)
		}

		contentType := http.DetectContentType(fileData)

		requestBody = map[string]interface{}{
			"contents": []interface{}{
				map[string]interface{}{
					"parts": []interface{}{
						map[string]interface{}{
							"text": prompt,
						},
						map[string]interface{}{
							"inline_data": map[string]interface{}{
								"mime_type": contentType,
								"data":      base64.StdEncoding.EncodeToString(fileData),
							},
						},
					},
				},
			},
		}
	} else {
		requestBody = map[string]interface{}{
			"contents": []interface{}{
				map[string]interface{}{
					"parts": []interface{}{
						map[string]string{"text": prompt},
					},
				},
			},
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var geminiResp GeminiResponse
	err = json.Unmarshal(body, &geminiResp)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response body: %w", err)
	}

	if len(geminiResp.Candidates) == 0 {
		return "No response from AI", nil
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func analyzePrescriptionHandler(w http.ResponseWriter, r *http.Request) {
	username, _, _ := getLoggedInUser(r)
	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	file, _, err := r.FormFile("prescription")
	if err != nil {
		http.Error(w, "Error uploading file", http.StatusBadRequest)
		return
	}

	prompt := `Analyze this prescription image and provide the following information in JSON format:
	1. List of medicines with their dosages
	2. Purpose/disease for each medicine
	3. Usage instructions
	4. Any warnings or contraindications
	5. Validate if the dosage seems appropriate (flag if suspicious)
	Format the response as a proper JSON object.`

	analysis, err := askGemini(prompt, file)
	if err != nil {
		http.Error(w, "AI service error", http.StatusInternalServerError)
		return
	}

	// Save prescription to database
	prescription := Prescription{
		PatientID:   username,
		Analysis:    analysis,
		UploadDate:  time.Now(),
	}

	_, err = prescriptionsColl.InsertOne(context.Background(), prescription)
	if err != nil {
		log.Printf("Error saving prescription: %v", err)
	}

	json.NewEncoder(w).Encode(map[string]string{"analysis": analysis})
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	username, role, _ := getLoggedInUser(r)
	data := PageData{User: username, Role: role}
	templates.ExecuteTemplate(w, "home.html", data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.ExecuteTemplate(w, "register.html", nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	role := "patient" // Only patient role is allowed now

	// Check if username already exists
	var existingUser User
	err := usersColl.FindOne(context.Background(), bson.M{"username": username}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "Username already exists", http.StatusBadRequest)
		return
	}

	// Create new user
	user := User{
		Username: username,
		Password: hashPassword(password),
		Role:     role,
	}

	_, err = usersColl.InsertOne(context.Background(), user)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user User
	err := usersColl.FindOne(context.Background(), bson.M{
		"username": username,
		"password": hashPassword(password),
	}).Decode(&user)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "username",
		Value:   user.Username,
		Expires: time.Now().Add(24 * time.Hour),
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "role",
		Value:   user.Role,
		Expires: time.Now().Add(24 * time.Hour),
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "username",
		Value:   "",
		Expires: time.Now(),
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "role",
		Value:   "",
		Expires: time.Now(),
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getLoggedInUser(r *http.Request) (string, string, bool) {
	usernameCookie, err1 := r.Cookie("username")
	roleCookie, err2 := r.Cookie("role")
	if err1 != nil || err2 != nil {
		return "", "", false
	}
	return usernameCookie.Value, roleCookie.Value, true
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	username, role, loggedIn := getLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user's prescriptions
	cursor, err := prescriptionsColl.Find(context.Background(), bson.M{"patient_id": username})
	if err != nil {
		log.Printf("Error fetching prescriptions: %v", err)
	}

	var prescriptions []Prescription
	if err = cursor.All(context.Background(), &prescriptions); err != nil {
		log.Printf("Error decoding prescriptions: %v", err)
	}

	data := PageData{
		User:         username,
		Role:         role,
		Prescriptions: prescriptions,
	}

	templates.ExecuteTemplate(w, "dashboard.html", data)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	_, _, loggedIn := getLoggedInUser(r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	prompt := "Act as a medical expert and answer in 200 characters. Answer this health query in a professional but understandable way: " + req.Message
	response, err := askGemini(prompt)
	if err != nil {
		http.Error(w, "AI service error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"response": response})
}

func predictDiseaseHandler(w http.ResponseWriter, r *http.Request) {
	_, _, loggedIn := getLoggedInUser(r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req DiseasePredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	prompt := fmt.Sprintf(`Act as a medical expert. Predict possible diseases based on these details:
	- Age: %s
	- Gender: %s
	- Symptoms: %s
	- Medical History: %s
	
	Provide potential diagnoses in order of likelihood, possible next steps, and when to seek urgent care.
	Use clear language without medical jargon. Answer in 450 characters. `, req.Age, req.Gender, req.Symptoms, req.MedicalHistory)

	response, err := askGemini(prompt)
	if err != nil {
		http.Error(w, "AI service error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"response": response})
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI(getMongoURI())
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database("medimate")
	usersColl = db.Collection("users")
	prescriptionsColl = db.Collection("prescriptions")

	// Create indexes
	_, err = usersColl.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/analyze-prescription", analyzePrescriptionHandler)
	http.HandleFunc("/chat", chatHandler)
	http.HandleFunc("/predict-disease", predictDiseaseHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
