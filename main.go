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
	"strings"
	"github.com/jung-kurt/gofpdf"
	"net/url"

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
	1. List of medicines with their:
	   - Name and dosage
	   - Purpose/disease
	   - Usage instructions
	   - Warnings or contraindications
	   - Dosage appropriateness (flag if suspicious)
	   - Generic alternatives (include name and approximate cost savings percentage)
	2. Dietary recommendations:
	   - List of foods to eat that can help with the condition
	   - List of foods to avoid that might interfere with the medication or condition
	3. Patient information (if available)
	4. Prescriber information
	5. Additional details like manufacturer, lot number, etc.

	Format the response as a proper JSON object with the following structure:
	{
		"patient_name": "...",
		"date": "...",
		"prescriber": "...",
		"medicines": [{
			"name": "...",
			"dosage": "...",
			"purpose": "...",
			"instructions": "...",
			"warnings": "...",
			"dosage_appropriate": "...",
			"generic_alternatives": [{
				"name": "...",
				"cost_saving": number
			}]
		}],
		"dietary_recommendations": {
			"foods_to_eat": ["..."],
			"foods_to_avoid": ["..."]
		},
		"manufacturer": "...",
		"lot_number": "...",
		"expiration_date": "..."
	}`

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

func getPrescriptionHandler(w http.ResponseWriter, r *http.Request) {
	username, _, loggedIn := getLoggedInUser(r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract prescription ID from URL
	prescriptionID := r.URL.Path[len("/prescription/"):]
	if prescriptionID == "" {
		http.Error(w, "Invalid prescription ID", http.StatusBadRequest)
		return
	}

	// Convert string ID to ObjectID
	objID, err := primitive.ObjectIDFromHex(prescriptionID)
	if err != nil {
		http.Error(w, "Invalid prescription ID format", http.StatusBadRequest)
		return
	}

	// Find prescription in database
	var prescription Prescription
	err = prescriptionsColl.FindOne(context.Background(), bson.M{
		"_id": objID,
		"patient_id": username, // Ensure user can only access their own prescriptions
	}).Decode(&prescription)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Prescription not found", http.StatusNotFound)
		} else {
			log.Printf("Error fetching prescription: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Clean the analysis string by removing markdown code block
	cleanAnalysis := prescription.Analysis
	cleanAnalysis = strings.TrimPrefix(cleanAnalysis, "```json\n")
	cleanAnalysis = strings.TrimSuffix(cleanAnalysis, "\n```")

	// Return prescription data as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"analysis": cleanAnalysis})
}

func downloadPrescriptionHandler(w http.ResponseWriter, r *http.Request) {
	username, _, loggedIn := getLoggedInUser(r)
	if !loggedIn {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract prescription ID from URL
	prescriptionID := r.URL.Path[len("/prescription/"):len(r.URL.Path)-len("/download")]
	if prescriptionID == "" {
		http.Error(w, "Invalid prescription ID", http.StatusBadRequest)
		return
	}

	// Convert string ID to ObjectID
	objID, err := primitive.ObjectIDFromHex(prescriptionID)
	if err != nil {
		http.Error(w, "Invalid prescription ID format", http.StatusBadRequest)
		return
	}

	// Find prescription in database
	var prescription Prescription
	err = prescriptionsColl.FindOne(context.Background(), bson.M{
		"_id": objID,
		"patient_id": username, // Ensure user can only access their own prescriptions
	}).Decode(&prescription)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Prescription not found", http.StatusNotFound)
		} else {
			log.Printf("Error fetching prescription: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Clean the analysis string by removing markdown code block
	cleanAnalysis := prescription.Analysis
	cleanAnalysis = strings.TrimPrefix(cleanAnalysis, "```json\n")
	cleanAnalysis = strings.TrimSuffix(cleanAnalysis, "\n```")

	// Parse the analysis JSON
	var analysis map[string]interface{}
	err = json.Unmarshal([]byte(cleanAnalysis), &analysis)
	if err != nil {
		log.Printf("Error parsing analysis data: %v", err)
		http.Error(w, "Error parsing analysis data", http.StatusInternalServerError)
		return
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Header
	pdf.Cell(190, 10, "MediMate Prescription Analysis Report")
	pdf.Ln(15)

	// Patient Information Section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Patient Information")
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(40, 8, "Date:")
	pdf.Cell(150, 8, prescription.UploadDate.Format("January 2, 2006 15:04:05"))
	pdf.Ln(8)
	
	pdf.Cell(40, 8, "Patient ID:")
	pdf.Cell(150, 8, prescription.PatientID)
	pdf.Ln(8)
	
	pdf.Cell(40, 8, "Patient Name:")
	pdf.Cell(150, 8, fmt.Sprintf("%v", analysis["patient_name"]))
	pdf.Ln(8)
	
	pdf.Cell(40, 8, "Prescriber:")
	pdf.Cell(150, 8, fmt.Sprintf("%v", analysis["prescriber"]))
	pdf.Ln(15)

	// Medicines Section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Prescribed Medicines")
	pdf.Ln(10)

	if medicines, ok := analysis["medicines"].([]interface{}); ok {
		for i, med := range medicines {
			if medicine, ok := med.(map[string]interface{}); ok {
				pdf.SetFont("Arial", "B", 11)
				pdf.Cell(190, 8, fmt.Sprintf("%d. %v", i+1, medicine["name"]))
				pdf.Ln(8)
				
				pdf.SetFont("Arial", "", 11)
				pdf.Cell(40, 8, "Dosage:")
				pdf.Cell(150, 8, fmt.Sprintf("%v", medicine["dosage"]))
				pdf.Ln(6)
				
				pdf.Cell(40, 8, "Purpose:")
				pdf.Cell(150, 8, fmt.Sprintf("%v", medicine["purpose"]))
				pdf.Ln(6)
				
				pdf.Cell(40, 8, "Instructions:")
				pdf.Cell(150, 8, fmt.Sprintf("%v", medicine["instructions"]))
				pdf.Ln(6)

				if warnings, exists := medicine["warnings"]; exists {
					pdf.Cell(40, 8, "Warnings:")
					// Use MultiCell for potentially long warning text
					currentX, currentY := pdf.GetXY()
					pdf.MultiCell(150, 8, fmt.Sprintf("%v", warnings), "", "", false)
					pdf.SetXY(currentX, currentY)
					pdf.Ln(8)
				}

				if appropriate, exists := medicine["dosage_appropriate"]; exists {
					pdf.Cell(40, 8, "Dosage Status:")
					pdf.Cell(150, 8, fmt.Sprintf("%v", appropriate))
					pdf.Ln(8)
				}

				// Add PharmEasy link
				if name, ok := medicine["name"].(string); ok && name != "" {
					pdf.Cell(40, 8, "Purchase Link:")
					pharmEasyLink := fmt.Sprintf("https://pharmeasy.in/search/all?name=%s", url.QueryEscape(name))
					pdf.SetTextColor(0, 0, 255) // Blue color for link
					pdf.Cell(150, 8, pharmEasyLink)
					pdf.SetTextColor(0, 0, 0) // Reset to black
					pdf.Ln(8)
				}

				// Add generic alternatives if available
				if generics, ok := medicine["generic_alternatives"].([]interface{}); ok && len(generics) > 0 {
					pdf.Cell(40, 8, "Generic Alternatives:")
					pdf.Ln(6)
					for _, gen := range generics {
						if generic, ok := gen.(map[string]interface{}); ok {
							pdf.Cell(20, 8, "•")
							pdf.Cell(130, 8, fmt.Sprintf("%v (%v%% cheaper)", generic["name"], generic["cost_saving"]))
							pdf.Ln(6)
						}
					}
					pdf.Ln(2)
				}

				pdf.Ln(4) // Space between medicines
			}
		}
	}

	// Add dietary recommendations if available
	if dietary, ok := analysis["dietary_recommendations"].(map[string]interface{}); ok {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 10, "Dietary Recommendations")
		pdf.Ln(10)

		// Foods to Eat
		if foods, ok := dietary["foods_to_eat"].([]interface{}); ok && len(foods) > 0 {
			pdf.SetFont("Arial", "B", 11)
			pdf.Cell(190, 8, "Foods to Eat:")
			pdf.Ln(8)
			pdf.SetFont("Arial", "", 11)
			for _, food := range foods {
				pdf.Cell(10, 8, "•")
				pdf.Cell(180, 8, fmt.Sprintf("%v", food))
				pdf.Ln(6)
			}
			pdf.Ln(4)
		}

		// Foods to Avoid
		if foods, ok := dietary["foods_to_avoid"].([]interface{}); ok && len(foods) > 0 {
			pdf.SetFont("Arial", "B", 11)
			pdf.Cell(190, 8, "Foods to Avoid:")
			pdf.Ln(8)
			pdf.SetFont("Arial", "", 11)
			for _, food := range foods {
				pdf.Cell(10, 8, "•")
				pdf.Cell(180, 8, fmt.Sprintf("%v", food))
				pdf.Ln(6)
			}
			pdf.Ln(4)
		}
	}

	// Additional Information Section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Additional Information")
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(40, 8, "Manufacturer:")
	pdf.Cell(150, 8, fmt.Sprintf("%v", analysis["manufacturer"]))
	pdf.Ln(8)
	
	pdf.Cell(40, 8, "Lot Number:")
	pdf.Cell(150, 8, fmt.Sprintf("%v", analysis["lot_number"]))
	pdf.Ln(8)
	
	pdf.Cell(40, 8, "Expiration Date:")
	pdf.Cell(150, 8, fmt.Sprintf("%v", analysis["expiration_date"]))
	pdf.Ln(8)

	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, fmt.Sprintf("Generated by MediMate on %s", time.Now().Format("January 2, 2006 15:04:05")))

	// Set response headers
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"prescription-analysis-%s.pdf\"", prescriptionID))
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Description", "File Transfer")

	// Write PDF to response
	err = pdf.Output(w)
	if err != nil {
		log.Printf("Error writing PDF: %v", err)
		http.Error(w, "Error generating PDF", http.StatusInternalServerError)
		return
	}
}

func deletePrescriptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get logged in user
	username, _, _ := getLoggedInUser(r)
	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get prescription ID from URL
	prescriptionID := r.URL.Query().Get("id")
	if prescriptionID == "" {
		http.Error(w, "Prescription ID is required", http.StatusBadRequest)
		return
	}

	// Convert string ID to ObjectID
	objID, err := primitive.ObjectIDFromHex(prescriptionID)
	if err != nil {
		http.Error(w, "Invalid prescription ID", http.StatusBadRequest)
		return
	}

	// Delete prescription
	filter := bson.M{
		"_id":       objID,
		"patientID": username, // Ensure user can only delete their own prescriptions
	}
	
	result, err := prescriptionsColl.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Error deleting prescription: %v", err)
		http.Error(w, "Error deleting prescription", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Prescription not found or unauthorized", http.StatusNotFound)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Prescription deleted successfully"})
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
	http.HandleFunc("/prescription/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/download") {
			downloadPrescriptionHandler(w, r)
		} else {
			getPrescriptionHandler(w, r)
		}
	})
	http.HandleFunc("/delete-prescription", deletePrescriptionHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
