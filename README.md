# MediMate - Smart Prescription Analysis

MediMate is an AI-powered prescription analysis tool that helps users understand their medical prescriptions better. It uses advanced AI technology to extract and analyze information from prescription images, providing detailed insights about medicines, dosages, and potential interactions.

## Deployed Link (Deployed on Google App Engine)

https://graphite-guard-462919-r1.de.r.appspot.com

## Problem Statement

*   **Medical Errors & Patient Safety**: Annually, **400,000 deaths** in India are attributed to adverse drug reactions (ADRs), with **5.2 million medical errors** reported. Prescription misinterpretation from poor handwriting leads to incorrect medication. (Source: NCBI, The Economic Times)
*   **Doctor Shortage**: India's doctor-population ratio of **1:834** (June 2022) results in long wait times and limited access to primary medical advice, especially in rural areas. (Source: The Economic Times, World Bank Data)
*   **Affordability of Medicines**: Branded medicines can be **5-22 times more expensive** than generics, leading to significant out-of-pocket costs for patients. This disparity highlights a potential saving of **INR 346.8 billion for statins** alone. (Source: PubMed, Cureus)

## Our Solution

MediMate addresses these issues with an AI-powered platform:

*   **AI Medical Chatbot**: Offers preliminary medical advice, easing doctor burden and addressing shortages.
*   **Prescription Analysis**: Digitizes handwritten prescriptions, reducing errors.
*   **Generic Alternatives**: Suggests cost-effective generic options, dosage, and dietary advice.
*   **Direct Purchase Links**: Streamlines medicine procurement via platforms like PharmEasy.

## How we utilized MongoDB

MongoDB was chosen for its key benefits in handling healthcare data:

*   **Flexible Data Model**: Adapts to diverse medical data without rigid schemas.
*   **Scalability**: Horizontally scales for growing users and analyses, ensuring high availability.
*   **Performance**: Enables rapid data retrieval for quick analysis and real-time AI responses.
*   **Rich Query Language**: Supports complex queries for historical data and search.
*   **Seamless Integration**: Efficiently integrates with Go (Golang) backend.

## How we utilized Google Cloud (specifically Google App Engine)

Google App Engine provides a robust and managed environment for MediMate:

*   **Scalability & Reliability**: Auto-scales based on demand, ensuring consistent performance.
*   **Managed Infrastructure**: Reduces operational overhead, focusing development on core features.
*   **Google Gemini AI Integration**: Leverages Gemini AI for advanced prescription analysis.
*   **Global Reach**: Deploys on Google's global network for low latency.
*   **Security & Compliance**: Benefits from robust security and HIPAA compliance for sensitive data.

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

## Future Enhancements

*   Integration with wearable health devices for proactive health monitoring.
*   Advanced AI models for personalized treatment plans and preventive care.
*   Multi-language support for broader accessibility.
*   Partnerships with more online pharmacies for wider medicine availability.

## Connect with Us

*   **LinkedIn**: [Link to LinkedIn Profile/Page]
*   **GitHub**: [Link to GitHub Profile/Organization]
*   **Email**: [Your Email Address]

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


## Screenshots



