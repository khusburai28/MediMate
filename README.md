# MediMate - Smart Prescription Analysis

MediMate is an AI-powered prescription analysis tool that helps users understand their medical prescriptions better. It uses advanced AI technology to extract and analyze information from prescription images, providing detailed insights about medicines, dosages, and potential interactions.

## Deployed Link (Deployed on Google App Engine)

https://graphite-guard-462919-r1.de.r.appspot.com

## Features

- **AI-Powered Prescription Analysis**
  - Upload prescription images for instant analysis
  - Extract medicine names, dosages, and instructions
  - Identify potential drug interactions and warnings
  - Validate dosage appropriateness

- **User Dashboard**
  - View all previous prescription analyses
  - Download detailed PDF reports
  - Direct links to purchase medicines on PharmEasy
  - Track prescription history

- **Security & Privacy**
  - Secure user authentication
  - Encrypted data storage
  - Private prescription access
  - HIPAA-compliant data handling

## Technology Stack

- **Backend**
  - Go (Golang) for server implementation
  - MongoDB for data storage
  - Google's Gemini AI for prescription analysis

- **Frontend**
  - HTML5, CSS3, JavaScript
  - Responsive design
  - Modern UI/UX principles

## Prerequisites

- Go 1.16 or higher
- MongoDB 4.4 or higher
- Google Cloud API credentials (for Gemini AI)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/khusburai28/medimate.git
   cd medimate
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Copy app.yaml.sample to app.yaml and update env for Google App Engine Deployment
   ```bash
   cp app.yaml.sample app.yaml
   # Edit app.yaml with your configuration
   ```

5. Start MongoDB:
   ```bash
   # Make sure MongoDB is running on your system
   ```

6. Run the application:
   ```bash
   go run main.go
   ```

7. Access the application:
   ```
   http://localhost:8080
   ```

## Environment Variables

Create a `.env` file with the following variables:

```env
MONGODB_URI=your_mongodb_connection_string
GOOGLE_API_KEY=your_google_api_key
SESSION_SECRET=your_session_secret
```

## Deployment on Google App Engine

```
# Navigate to folder
cd MediMate

# CLI Login
gcloud auth login

# Set Project
gcloud config set project PROJECT_ID_HERE

# Deploy
gcloud app deploy  ## Might fail on first attempt

# Set permission
gcloud projects add-iam-policy-binding PROJECT_ID_HERE --member="serviceAccount:PROJECT_ID_HERE@appspot.gserviceaccount.com"  --role="roles/storage.objectAdmin"

# Deploy again
gcloud app deploy

# Open Production Link
gcloud app browse
```

## API Endpoints

- `POST /analyze-prescription` - Upload and analyze a prescription
- `GET /prescription/:id` - View a specific prescription analysis
- `GET /prescription/:id/download` - Download prescription analysis as PDF
- `POST /chat` - Chat with AI about medical queries
- `POST /predict-disease` - Get disease predictions based on symptoms

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


## Screenshots



