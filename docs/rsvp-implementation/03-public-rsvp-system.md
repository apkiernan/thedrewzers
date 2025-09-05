# Phase 3: Public RSVP System

## Overview
This phase implements the guest-facing RSVP system, including the QR code landing page, RSVP form, and submission handling.

## Prerequisites
- Phase 1 & 2 completed
- Guest data imported
- QR codes generated

## Step 1: RSVP Page Handler

### 1.1 Create RSVP Handler
Create `internal/handlers/rsvp.go`:

```go
package handlers

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"
    
    "github.com/google/uuid"
    "github.com/apkiernan/thedrewzers/internal/db"
    "github.com/apkiernan/thedrewzers/internal/models"
    "github.com/apkiernan/thedrewzers/internal/views"
)

type RSVPHandler struct {
    guestRepo db.GuestRepository
    rsvpRepo  db.RSVPRepository
}

func NewRSVPHandler(guestRepo db.GuestRepository, rsvpRepo db.RSVPRepository) *RSVPHandler {
    return &RSVPHandler{
        guestRepo: guestRepo,
        rsvpRepo:  rsvpRepo,
    }
}

// HandleRSVPPage displays the RSVP form
func (h *RSVPHandler) HandleRSVPPage(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    if code == "" {
        // Show code entry form
        views.App(views.RSVPCodeEntry()).Render(r.Context(), w)
        return
    }
    
    // Look up guest by invitation code
    guest, err := h.guestRepo.GetGuestByInvitationCode(r.Context(), code)
    if err != nil {
        log.Printf("Failed to find guest with code %s: %v", code, err)
        views.App(views.RSVPNotFound()).Render(r.Context(), w)
        return
    }
    
    // Check for existing RSVP
    existingRSVP, _ := h.rsvpRepo.GetRSVP(r.Context(), guest.GuestID)
    
    // Render RSVP form
    views.App(views.RSVPForm(guest, existingRSVP)).Render(r.Context(), w)
}

// HandleRSVPLookup handles AJAX lookup requests
func (h *RSVPHandler) HandleRSVPLookup(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Code string `json:"code"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    guest, err := h.guestRepo.GetGuestByInvitationCode(r.Context(), req.Code)
    if err != nil {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid invitation code",
        })
        return
    }
    
    existingRSVP, _ := h.rsvpRepo.GetRSVP(r.Context(), guest.GuestID)
    
    response := map[string]interface{}{
        "guest":         guest,
        "existing_rsvp": existingRSVP,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// HandleRSVPSubmit processes RSVP submissions
func (h *RSVPHandler) HandleRSVPSubmit(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Parse form data
    var req models.RSVPRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Validate guest
    guest, err := h.guestRepo.GetGuestByInvitationCode(r.Context(), req.InvitationCode)
    if err != nil {
        http.Error(w, "Invalid invitation code", http.StatusBadRequest)
        return
    }
    
    // Validate party size
    if req.PartySize > guest.MaxPartySize {
        http.Error(w, "Party size exceeds maximum allowed", http.StatusBadRequest)
        return
    }
    
    // Create or update RSVP
    rsvp := &models.RSVP{
        RSVPID:              uuid.New().String(),
        GuestID:             guest.GuestID,
        Attending:           req.Attending,
        PartySize:           req.PartySize,
        AttendeeNames:       req.AttendeeNames,
        DietaryRestrictions: req.DietaryRestrictions,
        SpecialRequests:     req.SpecialRequests,
        SubmittedAt:         time.Now(),
        UpdatedAt:           time.Now(),
        IPAddress:           getClientIP(r),
        UserAgent:           r.UserAgent(),
    }
    
    // Check if updating existing RSVP
    existingRSVP, _ := h.rsvpRepo.GetRSVP(r.Context(), guest.GuestID)
    if existingRSVP != nil {
        rsvp.RSVPID = existingRSVP.RSVPID
        err = h.rsvpRepo.UpdateRSVP(r.Context(), rsvp)
    } else {
        err = h.rsvpRepo.CreateRSVP(r.Context(), rsvp)
    }
    
    if err != nil {
        log.Printf("Failed to save RSVP: %v", err)
        http.Error(w, "Failed to save RSVP", http.StatusInternalServerError)
        return
    }
    
    // Send confirmation email (if email service is configured)
    if guest.Email != "" {
        go h.sendConfirmationEmail(guest, rsvp)
    }
    
    // Return success response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "RSVP submitted successfully",
        "rsvp":    rsvp,
    })
}

func getClientIP(r *http.Request) string {
    // Check X-Forwarded-For header (from CloudFront)
    xff := r.Header.Get("X-Forwarded-For")
    if xff != "" {
        return xff
    }
    return r.RemoteAddr
}

func (h *RSVPHandler) sendConfirmationEmail(guest *models.Guest, rsvp *models.RSVP) {
    // TODO: Implement email sending
    log.Printf("Would send confirmation email to %s for RSVP %s", guest.Email, rsvp.RSVPID)
}
```

### 1.2 Create Request/Response Types
Add to `internal/models/rsvp.go`:

```go
type RSVPRequest struct {
    GuestID             string   `json:"guest_id"`
    InvitationCode      string   `json:"invitation_code"`
    Attending           bool     `json:"attending"`
    PartySize           int      `json:"party_size"`
    AttendeeNames       []string `json:"attendee_names"`
    DietaryRestrictions []string `json:"dietary_restrictions"`
    SpecialRequests     string   `json:"special_requests"`
}

type RSVPResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    RSVP    *RSVP  `json:"rsvp,omitempty"`
    Error   string `json:"error,omitempty"`
}
```

## Step 2: RSVP Views

### 2.1 Create RSVP Form Template
Create `internal/views/rsvp_form.templ`:

```templ
package views

import (
    "strconv"
    "github.com/apkiernan/thedrewzers/internal/models"
)

templ RSVPForm(guest *models.Guest, existingRSVP *models.RSVP) {
    <div class="min-h-screen bg-white py-12">
        <div class="max-w-2xl mx-auto px-6">
            <div class="text-center mb-12">
                <h1 class="script text-5xl md:text-6xl text-blue-300 mb-6 font-extralight">RSVP</h1>
                <p class="text-xl text-gray-600">Hello, { guest.PrimaryGuest }!</p>
                <p class="text-gray-500 mt-2">We'd love to celebrate with you</p>
            </div>
            
            <form id="rsvp-form" class="space-y-8" data-guest-id={ guest.GuestID } data-code={ guest.InvitationCode }>
                // Attending choice
                <div class="text-center">
                    <p class="text-gray-700 mb-4">Will you be joining us?</p>
                    <div class="flex justify-center space-x-8">
                        <label class="cursor-pointer">
                            <input 
                                type="radio" 
                                name="attending" 
                                value="yes" 
                                class="sr-only peer"
                                if existingRSVP != nil && existingRSVP.Attending {
                                    checked
                                }
                                onchange="toggleAttendingDetails(true)"
                            />
                            <div class="px-8 py-4 border-2 border-gray-300 rounded-lg peer-checked:border-green-500 peer-checked:bg-green-50 transition-colors">
                                <span class="text-lg">Joyfully Accept</span>
                            </div>
                        </label>
                        
                        <label class="cursor-pointer">
                            <input 
                                type="radio" 
                                name="attending" 
                                value="no" 
                                class="sr-only peer"
                                if existingRSVP != nil && !existingRSVP.Attending {
                                    checked
                                }
                                onchange="toggleAttendingDetails(false)"
                            />
                            <div class="px-8 py-4 border-2 border-gray-300 rounded-lg peer-checked:border-red-500 peer-checked:bg-red-50 transition-colors">
                                <span class="text-lg">Regretfully Decline</span>
                            </div>
                        </label>
                    </div>
                </div>
                
                // Attending details
                <div id="attending-details" class={
                    if existingRSVP == nil || !existingRSVP.Attending { "hidden" } else { "" }
                }>
                    // Party size
                    <div>
                        <label class="block text-gray-700 mb-2">
                            How many will be attending?
                            <span class="text-sm text-gray-500">(Maximum { strconv.Itoa(guest.MaxPartySize) })</span>
                        </label>
                        <select 
                            name="party_size" 
                            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-300 focus:border-transparent"
                            onchange="updateAttendeeNames(this.value)"
                        >
                            @for i := 1; i <= guest.MaxPartySize; i++ {
                                <option 
                                    value={ strconv.Itoa(i) }
                                    if existingRSVP != nil && existingRSVP.PartySize == i { selected }
                                >
                                    { strconv.Itoa(i) } { if i == 1 { "Guest" } else { "Guests" } }
                                </option>
                            }
                        </select>
                    </div>
                    
                    // Attendee names
                    <div>
                        <label class="block text-gray-700 mb-2">
                            Names of attendees
                            <span class="text-sm text-gray-500">(Including yourself)</span>
                        </label>
                        <div id="attendee-names" class="space-y-2">
                            @if existingRSVP != nil {
                                @for i, name := range existingRSVP.AttendeeNames {
                                    <input 
                                        type="text" 
                                        name="attendee_names[]" 
                                        value={ name }
                                        placeholder={ "Guest " + strconv.Itoa(i+1) + " name" }
                                        class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-300 focus:border-transparent"
                                    />
                                }
                            } else {
                                <input 
                                    type="text" 
                                    name="attendee_names[]" 
                                    placeholder="Guest 1 name"
                                    value={ guest.PrimaryGuest }
                                    class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-300 focus:border-transparent"
                                />
                            }
                        </div>
                    </div>
                    
                    // Dietary restrictions
                    <div>
                        <label class="block text-gray-700 mb-2">
                            Dietary restrictions or allergies
                            <span class="text-sm text-gray-500">(Optional)</span>
                        </label>
                        <textarea 
                            name="dietary_restrictions" 
                            rows="3"
                            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-300 focus:border-transparent"
                            placeholder="Please let us know about any dietary needs..."
                        >{ if existingRSVP != nil { existingRSVP.SpecialRequests } }</textarea>
                    </div>
                </div>
                
                // Submit button
                <div class="text-center pt-8">
                    <button 
                        type="submit"
                        class="px-12 py-4 bg-blue-300 text-white rounded-lg hover:bg-blue-400 transition-colors text-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                        id="submit-button"
                    >
                        Submit RSVP
                    </button>
                </div>
            </form>
        </div>
    </div>
    
    <script src="/static/js/rsvp.js"></script>
}

templ RSVPCodeEntry() {
    <div class="min-h-screen bg-white py-12">
        <div class="max-w-md mx-auto px-6">
            <div class="text-center mb-12">
                <h1 class="script text-5xl md:text-6xl text-blue-300 mb-6 font-extralight">RSVP</h1>
                <p class="text-gray-600">Please enter your invitation code</p>
            </div>
            
            <form id="code-form" class="space-y-6">
                <div>
                    <label class="block text-gray-700 mb-2">Invitation Code</label>
                    <input 
                        type="text" 
                        name="code" 
                        class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-300 focus:border-transparent text-center uppercase"
                        placeholder="ABC12345"
                        maxlength="8"
                        required
                    />
                </div>
                
                <button 
                    type="submit"
                    class="w-full px-6 py-3 bg-blue-300 text-white rounded-lg hover:bg-blue-400 transition-colors"
                >
                    Continue
                </button>
            </form>
            
            <p class="text-center text-sm text-gray-500 mt-8">
                Your invitation code can be found on your invitation
            </p>
        </div>
    </div>
    
    <script>
        document.getElementById('code-form').addEventListener('submit', function(e) {
            e.preventDefault();
            const code = e.target.code.value.toUpperCase();
            window.location.href = '/rsvp?code=' + code;
        });
    </script>
}

templ RSVPNotFound() {
    <div class="min-h-screen bg-white py-12">
        <div class="max-w-md mx-auto px-6 text-center">
            <h1 class="script text-5xl md:text-6xl text-blue-300 mb-6 font-extralight">Oops!</h1>
            <p class="text-gray-600 mb-8">We couldn't find that invitation code.</p>
            <p class="text-gray-500 mb-8">Please double-check the code on your invitation and try again.</p>
            <a 
                href="/rsvp"
                class="inline-block px-6 py-3 bg-blue-300 text-white rounded-lg hover:bg-blue-400 transition-colors"
            >
                Try Again
            </a>
        </div>
    </div>
}

templ RSVPSuccess() {
    <div class="min-h-screen bg-white py-12">
        <div class="max-w-md mx-auto px-6 text-center">
            <div class="mb-8">
                <svg class="w-24 h-24 text-green-500 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
            </div>
            <h1 class="script text-5xl md:text-6xl text-blue-300 mb-6 font-extralight">Thank You!</h1>
            <p class="text-gray-600 mb-4">Your RSVP has been received.</p>
            <p class="text-gray-500 mb-8">We've sent a confirmation to your email address.</p>
            <a 
                href="/"
                class="inline-block px-6 py-3 bg-blue-300 text-white rounded-lg hover:bg-blue-400 transition-colors"
            >
                Return to Homepage
            </a>
        </div>
    </div>
}
```

### 2.2 Create RSVP JavaScript
Create `static/js/rsvp.js`:

```javascript
// Toggle attending details visibility
function toggleAttendingDetails(show) {
    const details = document.getElementById('attending-details');
    if (show) {
        details.classList.remove('hidden');
    } else {
        details.classList.add('hidden');
    }
}

// Update attendee name inputs based on party size
function updateAttendeeNames(partySize) {
    const container = document.getElementById('attendee-names');
    const currentInputs = container.querySelectorAll('input');
    const currentCount = currentInputs.length;
    const newCount = parseInt(partySize);
    
    if (newCount > currentCount) {
        // Add more inputs
        for (let i = currentCount; i < newCount; i++) {
            const input = document.createElement('input');
            input.type = 'text';
            input.name = 'attendee_names[]';
            input.placeholder = `Guest ${i + 1} name`;
            input.className = 'w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-300 focus:border-transparent';
            container.appendChild(input);
        }
    } else if (newCount < currentCount) {
        // Remove extra inputs
        for (let i = currentCount - 1; i >= newCount; i--) {
            container.removeChild(currentInputs[i]);
        }
    }
}

// Handle form submission
document.getElementById('rsvp-form')?.addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const form = e.target;
    const submitButton = document.getElementById('submit-button');
    const guestId = form.dataset.guestId;
    const code = form.dataset.code;
    
    // Disable submit button
    submitButton.disabled = true;
    submitButton.textContent = 'Submitting...';
    
    // Gather form data
    const formData = new FormData(form);
    const attending = formData.get('attending') === 'yes';
    
    const data = {
        guest_id: guestId,
        invitation_code: code,
        attending: attending,
        party_size: attending ? parseInt(formData.get('party_size')) : 0,
        attendee_names: attending ? formData.getAll('attendee_names[]').filter(n => n) : [],
        dietary_restrictions: attending ? formData.get('dietary_restrictions')?.split('\n').filter(n => n) : [],
        special_requests: attending ? formData.get('dietary_restrictions') : ''
    };
    
    try {
        const response = await fetch('/api/rsvp/submit', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });
        
        const result = await response.json();
        
        if (response.ok && result.success) {
            // Show success message
            window.location.href = '/rsvp/success';
        } else {
            throw new Error(result.error || 'Failed to submit RSVP');
        }
    } catch (error) {
        console.error('RSVP submission error:', error);
        alert('Sorry, there was an error submitting your RSVP. Please try again.');
        submitButton.disabled = false;
        submitButton.textContent = 'Submit RSVP';
    }
});

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    // Set initial state based on existing RSVP
    const attendingRadio = document.querySelector('input[name="attending"]:checked');
    if (attendingRadio) {
        toggleAttendingDetails(attendingRadio.value === 'yes');
    }
    
    // Initialize party size if set
    const partySize = document.querySelector('select[name="party_size"]');
    if (partySize && partySize.value) {
        updateAttendeeNames(partySize.value);
    }
});
```

## Step 3: Wire Up Routes

### 3.1 Update Main Lambda Handler
Update `cmd/lambda/main.go`:

```go
package main

import (
    "context"
    "os"
    
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
    
    "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
    "github.com/apkiernan/thedrewzers/internal/handlers"
)

var httpAdapter *httpadapter.HandlerAdapter

func init() {
    // Initialize AWS config
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        panic(err)
    }
    
    // Initialize DynamoDB client
    dynamoClient := dynamodb.NewFromConfig(cfg)
    
    // Initialize repositories
    guestRepo := dynamodb.NewGuestRepository(
        dynamoClient, 
        os.Getenv("GUESTS_TABLE"),
    )
    rsvpRepo := dynamodb.NewRSVPRepository(
        dynamoClient,
        os.Getenv("RSVPS_TABLE"),
    )
    
    // Initialize handlers
    rsvpHandler := handlers.NewRSVPHandler(guestRepo, rsvpRepo)
    
    // Setup routes
    mux := http.NewServeMux()
    
    // Existing routes
    mux.HandleFunc("/", handlers.HandleHomePage)
    
    // RSVP routes
    mux.HandleFunc("/rsvp", rsvpHandler.HandleRSVPPage)
    mux.HandleFunc("/rsvp/success", func(w http.ResponseWriter, r *http.Request) {
        views.App(views.RSVPSuccess()).Render(r.Context(), w)
    })
    mux.HandleFunc("/api/rsvp/lookup", rsvpHandler.HandleRSVPLookup)
    mux.HandleFunc("/api/rsvp/submit", rsvpHandler.HandleRSVPSubmit)
    
    // Create adapter
    httpAdapter = httpadapter.New(mux)
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    return httpAdapter.ProxyWithContext(ctx, req)
}

func main() {
    lambda.Start(handler)
}
```

## Step 4: Testing

### 4.1 Local Testing Setup
Create `cmd/local/main.go`:

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    
    dbRepo "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
    "github.com/apkiernan/thedrewzers/internal/handlers"
)

func main() {
    // Setup AWS config
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Fatal(err)
    }
    
    // Initialize DynamoDB
    dynamoClient := dynamodb.NewFromConfig(cfg)
    
    // Initialize repositories
    guestRepo := dbRepo.NewGuestRepository(
        dynamoClient,
        os.Getenv("GUESTS_TABLE"),
    )
    rsvpRepo := dbRepo.NewRSVPRepository(
        dynamoClient,
        os.Getenv("RSVPS_TABLE"),
    )
    
    // Initialize handlers
    rsvpHandler := handlers.NewRSVPHandler(guestRepo, rsvpRepo)
    
    // Setup routes
    mux := http.NewServeMux()
    mux.HandleFunc("/", handlers.HandleHomePage)
    mux.HandleFunc("/rsvp", rsvpHandler.HandleRSVPPage)
    mux.HandleFunc("/rsvp/success", func(w http.ResponseWriter, r *http.Request) {
        views.App(views.RSVPSuccess()).Render(r.Context(), w)
    })
    mux.HandleFunc("/api/rsvp/lookup", rsvpHandler.HandleRSVPLookup)
    mux.HandleFunc("/api/rsvp/submit", rsvpHandler.HandleRSVPSubmit)
    
    // Serve static files
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
    
    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

### 4.2 Test RSVP Flow
```bash
# Set environment variables
export GUESTS_TABLE=wedding-guests
export RSVPS_TABLE=wedding-rsvps

# Run local server
go run cmd/local/main.go

# Test with a valid invitation code
open http://localhost:8080/rsvp?code=ABC12345
```

### 4.3 Test API Endpoints
```bash
# Test lookup
curl -X POST http://localhost:8080/api/rsvp/lookup \
  -H "Content-Type: application/json" \
  -d '{"code":"ABC12345"}'

# Test RSVP submission
curl -X POST http://localhost:8080/api/rsvp/submit \
  -H "Content-Type: application/json" \
  -d '{
    "guest_id": "123",
    "invitation_code": "ABC12345",
    "attending": true,
    "party_size": 2,
    "attendee_names": ["John Smith", "Jane Smith"],
    "dietary_restrictions": ["Vegetarian"],
    "special_requests": "No nuts please"
  }'
```

## Step 5: Mobile Optimization

### 5.1 Add Viewport Meta Tag
Update `internal/views/app.templ`:

```templ
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no"/>
```

### 5.2 Test QR Code Scanning
1. Generate a test QR code pointing to your local server
2. Test on various devices (iPhone, Android)
3. Verify form is mobile-friendly

## Next Steps
- Phase 4: Admin authentication system
- Add email confirmation service
- Implement calendar invite generation

## Deployment Checklist
- [ ] Update Lambda function with RSVP handlers
- [ ] Deploy new static assets (JS files)
- [ ] Test QR code scanning on production
- [ ] Monitor DynamoDB for successful submissions