# Phase 5: Admin Dashboard Implementation

## Overview
This phase implements the admin dashboard with RSVP statistics, guest management, and export functionality.

## Prerequisites
- Phase 1-4 completed
- Admin authentication working
- Admin user created

## Step 1: Dashboard Statistics

### 1.1 Create Dashboard Types
Create `internal/models/dashboard.go`:

```go
package models

type DashboardStats struct {
    TotalInvited       int                    `json:"total_invited"`
    TotalResponses     int                    `json:"total_responses"`
    TotalAttending     int                    `json:"total_attending"`
    TotalDeclined      int                    `json:"total_declined"`
    TotalPending       int                    `json:"total_pending"`
    ResponseRate       float64                `json:"response_rate"`
    AttendingGuests    int                    `json:"attending_guests"`
    DietaryBreakdown   map[string]int         `json:"dietary_breakdown"`
    RecentRSVPs        []RecentRSVP           `json:"recent_rsvps"`
}

type RecentRSVP struct {
    GuestName    string    `json:"guest_name"`
    Attending    bool      `json:"attending"`
    PartySize    int       `json:"party_size"`
    SubmittedAt  time.Time `json:"submitted_at"`
}

type GuestWithRSVP struct {
    Guest *Guest `json:"guest"`
    RSVP  *RSVP  `json:"rsvp,omitempty"`
}

type ExportData struct {
    Headers []string   `json:"headers"`
    Rows    [][]string `json:"rows"`
}
```

### 1.2 Create Statistics Service
Create `internal/services/stats_service.go`:

```go
package services

import (
    "context"
    "strings"
    "time"
    
    "github.com/apkiernan/thedrewzers/internal/db"
    "github.com/apkiernan/thedrewzers/internal/models"
)

type StatsService struct {
    guestRepo db.GuestRepository
    rsvpRepo  db.RSVPRepository
}

func NewStatsService(guestRepo db.GuestRepository, rsvpRepo db.RSVPRepository) *StatsService {
    return &StatsService{
        guestRepo: guestRepo,
        rsvpRepo:  rsvpRepo,
    }
}

func (s *StatsService) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
    // Get all guests
    guests, err := s.guestRepo.ListGuests(ctx)
    if err != nil {
        return nil, err
    }
    
    // Get all RSVPs
    rsvps, err := s.rsvpRepo.ListRSVPs(ctx)
    if err != nil {
        return nil, err
    }
    
    // Create RSVP map for quick lookup
    rsvpMap := make(map[string]*models.RSVP)
    for _, rsvp := range rsvps {
        rsvpMap[rsvp.GuestID] = rsvp
    }
    
    // Calculate statistics
    stats := &models.DashboardStats{
        TotalInvited:     len(guests),
        TotalResponses:   len(rsvps),
        DietaryBreakdown: make(map[string]int),
        RecentRSVPs:      make([]models.RecentRSVP, 0),
    }
    
    // Count attending/declined and dietary restrictions
    for _, rsvp := range rsvps {
        if rsvp.Attending {
            stats.TotalAttending++
            stats.AttendingGuests += rsvp.PartySize
            
            // Count dietary restrictions
            for _, restriction := range rsvp.DietaryRestrictions {
                normalized := strings.TrimSpace(strings.ToLower(restriction))
                if normalized != "" {
                    stats.DietaryBreakdown[normalized]++
                }
            }
        } else {
            stats.TotalDeclined++
        }
    }
    
    stats.TotalPending = stats.TotalInvited - stats.TotalResponses
    
    if stats.TotalInvited > 0 {
        stats.ResponseRate = float64(stats.TotalResponses) / float64(stats.TotalInvited) * 100
    }
    
    // Get recent RSVPs (last 10)
    recentRSVPs := s.getRecentRSVPs(rsvps, guests, 10)
    stats.RecentRSVPs = recentRSVPs
    
    return stats, nil
}

func (s *StatsService) getRecentRSVPs(rsvps []*models.RSVP, guests []*models.Guest, limit int) []models.RecentRSVP {
    // Create guest map
    guestMap := make(map[string]*models.Guest)
    for _, guest := range guests {
        guestMap[guest.GuestID] = guest
    }
    
    // Sort RSVPs by submission time (newest first)
    // In production, you'd use a proper sorting algorithm
    recent := make([]models.RecentRSVP, 0, limit)
    
    for i := len(rsvps) - 1; i >= 0 && len(recent) < limit; i-- {
        rsvp := rsvps[i]
        if guest, ok := guestMap[rsvp.GuestID]; ok {
            recent = append(recent, models.RecentRSVP{
                GuestName:   guest.PrimaryGuest,
                Attending:   rsvp.Attending,
                PartySize:   rsvp.PartySize,
                SubmittedAt: rsvp.SubmittedAt,
            })
        }
    }
    
    return recent
}

func (s *StatsService) GetGuestsWithRSVPs(ctx context.Context) ([]*models.GuestWithRSVP, error) {
    guests, err := s.guestRepo.ListGuests(ctx)
    if err != nil {
        return nil, err
    }
    
    rsvps, err := s.rsvpRepo.ListRSVPs(ctx)
    if err != nil {
        return nil, err
    }
    
    // Create RSVP map
    rsvpMap := make(map[string]*models.RSVP)
    for _, rsvp := range rsvps {
        rsvpMap[rsvp.GuestID] = rsvp
    }
    
    // Combine data
    result := make([]*models.GuestWithRSVP, 0, len(guests))
    for _, guest := range guests {
        result = append(result, &models.GuestWithRSVP{
            Guest: guest,
            RSVP:  rsvpMap[guest.GuestID],
        })
    }
    
    return result, nil
}
```

## Step 2: Admin Dashboard Handlers

### 2.1 Create Dashboard Handler
Create `internal/handlers/admin_dashboard.go`:

```go
package handlers

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"
    
    "github.com/apkiernan/thedrewzers/internal/auth"
    "github.com/apkiernan/thedrewzers/internal/middleware"
    "github.com/apkiernan/thedrewzers/internal/services"
    "github.com/apkiernan/thedrewzers/internal/views"
)

type AdminDashboardHandler struct {
    statsService *services.StatsService
}

func NewAdminDashboardHandler(statsService *services.StatsService) *AdminDashboardHandler {
    return &AdminDashboardHandler{
        statsService: statsService,
    }
}

func (h *AdminDashboardHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
    // Get user info from context
    claims := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)
    
    // Get dashboard stats
    stats, err := h.statsService.GetDashboardStats(r.Context())
    if err != nil {
        http.Error(w, "Failed to load dashboard", http.StatusInternalServerError)
        return
    }
    
    // Render dashboard
    views.AdminDashboard(claims.Name, stats).Render(r.Context(), w)
}

func (h *AdminDashboardHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
    stats, err := h.statsService.GetDashboardStats(r.Context())
    if err != nil {
        http.Error(w, "Failed to get stats", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

func (h *AdminDashboardHandler) HandleGuests(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)
    
    // Get all guests with RSVPs
    guestsWithRSVPs, err := h.statsService.GetGuestsWithRSVPs(r.Context())
    if err != nil {
        http.Error(w, "Failed to load guests", http.StatusInternalServerError)
        return
    }
    
    // Render guest list
    views.AdminGuestList(claims.Name, guestsWithRSVPs).Render(r.Context(), w)
}

func (h *AdminDashboardHandler) HandleRSVPs(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)
    
    // Get all guests with RSVPs, filtered to only those who responded
    guestsWithRSVPs, err := h.statsService.GetGuestsWithRSVPs(r.Context())
    if err != nil {
        http.Error(w, "Failed to load RSVPs", http.StatusInternalServerError)
        return
    }
    
    // Filter to only guests with RSVPs
    respondedGuests := make([]*models.GuestWithRSVP, 0)
    for _, gwRSVP := range guestsWithRSVPs {
        if gwRSVP.RSVP != nil {
            respondedGuests = append(respondedGuests, gwRSVP)
        }
    }
    
    views.AdminRSVPList(claims.Name, respondedGuests).Render(r.Context(), w)
}

func (h *AdminDashboardHandler) HandleExportRSVPs(w http.ResponseWriter, r *http.Request) {
    // Get all guests with RSVPs
    guestsWithRSVPs, err := h.statsService.GetGuestsWithRSVPs(r.Context())
    if err != nil {
        http.Error(w, "Failed to export data", http.StatusInternalServerError)
        return
    }
    
    // Set CSV headers
    w.Header().Set("Content-Type", "text/csv")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=rsvps_%s.csv", 
        time.Now().Format("2006-01-02")))
    
    // Create CSV writer
    writer := csv.NewWriter(w)
    defer writer.Flush()
    
    // Write headers
    headers := []string{
        "Primary Guest",
        "Email",
        "Invitation Code",
        "Max Party Size",
        "RSVP Status",
        "Attending",
        "Party Size",
        "Attendee Names",
        "Dietary Restrictions",
        "Special Requests",
        "Submitted At",
    }
    writer.Write(headers)
    
    // Write data rows
    for _, gwRSVP := range guestsWithRSVPs {
        row := []string{
            gwRSVP.Guest.PrimaryGuest,
            gwRSVP.Guest.Email,
            gwRSVP.Guest.InvitationCode,
            strconv.Itoa(gwRSVP.Guest.MaxPartySize),
        }
        
        if gwRSVP.RSVP != nil {
            row = append(row,
                "Responded",
                strconv.FormatBool(gwRSVP.RSVP.Attending),
                strconv.Itoa(gwRSVP.RSVP.PartySize),
                strings.Join(gwRSVP.RSVP.AttendeeNames, "; "),
                strings.Join(gwRSVP.RSVP.DietaryRestrictions, "; "),
                gwRSVP.RSVP.SpecialRequests,
                gwRSVP.RSVP.SubmittedAt.Format("2006-01-02 15:04:05"),
            )
        } else {
            row = append(row,
                "Pending",
                "",
                "",
                "",
                "",
                "",
                "",
            )
        }
        
        writer.Write(row)
    }
}
```

## Step 3: Admin Dashboard Views

### 3.1 Create Admin Layout
Create `internal/views/admin_layout.templ`:

```templ
package views

import "github.com/apkiernan/thedrewzers/internal/middleware"

templ AdminLayout(userName string, activePage string, content templ.Component) {
    <!DOCTYPE html>
    <html lang="en">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            <title>{ activePage } - Wedding Admin</title>
            <link href="/static/css/tailwind.css" rel="stylesheet"/>
            <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
        </head>
        <body class="bg-gray-50">
            <nav class="bg-white shadow-sm border-b border-gray-200">
                <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div class="flex justify-between h-16">
                        <div class="flex">
                            <div class="flex-shrink-0 flex items-center">
                                <h1 class="text-xl font-semibold text-gray-800">Wedding Admin</h1>
                            </div>
                            <div class="hidden sm:ml-8 sm:flex sm:space-x-8">
                                <a 
                                    href="/dashboard" 
                                    class={ "inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium",
                                        if activePage == "Dashboard" { 
                                            "border-indigo-500 text-gray-900" 
                                        } else { 
                                            "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700" 
                                        }
                                    }
                                >
                                    Dashboard
                                </a>
                                <a 
                                    href="/guests" 
                                    class={ "inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium",
                                        if activePage == "Guests" { 
                                            "border-indigo-500 text-gray-900" 
                                        } else { 
                                            "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700" 
                                        }
                                    }
                                >
                                    Guests
                                </a>
                                <a 
                                    href="/rsvps" 
                                    class={ "inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium",
                                        if activePage == "RSVPs" { 
                                            "border-indigo-500 text-gray-900" 
                                        } else { 
                                            "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700" 
                                        }
                                    }
                                >
                                    RSVPs
                                </a>
                            </div>
                        </div>
                        <div class="flex items-center">
                            <span class="text-sm text-gray-500 mr-4">{ userName }</span>
                            <a 
                                href="/logout" 
                                class="text-sm text-gray-500 hover:text-gray-700"
                            >
                                Logout
                            </a>
                        </div>
                    </div>
                </div>
            </nav>
            <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
                @content
            </main>
        </body>
    </html>
}
```

### 3.2 Create Dashboard View
Create `internal/views/admin_dashboard.templ`:

```templ
package views

import (
    "fmt"
    "github.com/apkiernan/thedrewzers/internal/models"
)

templ AdminDashboard(userName string, stats *models.DashboardStats) {
    @AdminLayout(userName, "Dashboard", AdminDashboardContent(stats))
}

templ AdminDashboardContent(stats *models.DashboardStats) {
    <div class="px-4 sm:px-0">
        <h2 class="text-2xl font-bold text-gray-900 mb-8">RSVP Dashboard</h2>
        
        // Statistics Cards
        <div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
            @StatCard("Total Invited", fmt.Sprintf("%d", stats.TotalInvited), "text-gray-900", "bg-gray-100")
            @StatCard("Responses", fmt.Sprintf("%d (%.1f%%)", stats.TotalResponses, stats.ResponseRate), "text-blue-600", "bg-blue-100")
            @StatCard("Attending", fmt.Sprintf("%d guests", stats.AttendingGuests), "text-green-600", "bg-green-100")
            @StatCard("Declined", fmt.Sprintf("%d", stats.TotalDeclined), "text-red-600", "bg-red-100")
        </div>
        
        // Charts Row
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
            // Response Status Chart
            <div class="bg-white p-6 rounded-lg shadow">
                <h3 class="text-lg font-medium text-gray-900 mb-4">Response Status</h3>
                <canvas id="responseChart" width="400" height="200"></canvas>
            </div>
            
            // Dietary Restrictions
            <div class="bg-white p-6 rounded-lg shadow">
                <h3 class="text-lg font-medium text-gray-900 mb-4">Dietary Restrictions</h3>
                if len(stats.DietaryBreakdown) > 0 {
                    <ul class="space-y-2">
                        @for restriction, count := range stats.DietaryBreakdown {
                            <li class="flex justify-between text-sm">
                                <span class="text-gray-600 capitalize">{ restriction }</span>
                                <span class="font-medium">{ fmt.Sprintf("%d", count) }</span>
                            </li>
                        }
                    </ul>
                } else {
                    <p class="text-gray-500 text-sm">No dietary restrictions reported yet</p>
                }
            </div>
        </div>
        
        // Recent RSVPs
        <div class="bg-white shadow rounded-lg">
            <div class="px-6 py-4 border-b border-gray-200">
                <h3 class="text-lg font-medium text-gray-900">Recent RSVPs</h3>
            </div>
            <div class="overflow-x-auto">
                <table class="min-w-full divide-y divide-gray-200">
                    <thead class="bg-gray-50">
                        <tr>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Guest</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Party Size</th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Submitted</th>
                        </tr>
                    </thead>
                    <tbody class="bg-white divide-y divide-gray-200">
                        if len(stats.RecentRSVPs) > 0 {
                            @for _, rsvp := range stats.RecentRSVPs {
                                <tr>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                                        { rsvp.GuestName }
                                    </td>
                                    <td class="px-6 py-4 whitespace-nowrap">
                                        if rsvp.Attending {
                                            <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                                                Attending
                                            </span>
                                        } else {
                                            <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-red-100 text-red-800">
                                                Declined
                                            </span>
                                        }
                                    </td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                        { fmt.Sprintf("%d", rsvp.PartySize) }
                                    </td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                        { rsvp.SubmittedAt.Format("Jan 2, 3:04 PM") }
                                    </td>
                                </tr>
                            }
                        } else {
                            <tr>
                                <td colspan="4" class="px-6 py-4 text-center text-sm text-gray-500">
                                    No RSVPs yet
                                </td>
                            </tr>
                        }
                    </tbody>
                </table>
            </div>
        </div>
        
        // Quick Actions
        <div class="mt-8 flex space-x-4">
            <a 
                href="/guests/add" 
                class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
            >
                Add Guest
            </a>
            <a 
                href="/rsvps/export" 
                class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
            >
                Export RSVPs
            </a>
        </div>
    </div>
    
    <script>
        // Response Status Chart
        const ctx = document.getElementById('responseChart').getContext('2d');
        new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['Attending', 'Declined', 'Pending'],
                datasets: [{
                    data: [
                        { fmt.Sprintf("%d", stats.TotalAttending) },
                        { fmt.Sprintf("%d", stats.TotalDeclined) },
                        { fmt.Sprintf("%d", stats.TotalPending) }
                    ],
                    backgroundColor: [
                        'rgb(34, 197, 94)',
                        'rgb(239, 68, 68)',
                        'rgb(156, 163, 175)'
                    ]
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom',
                    }
                }
            }
        });
    </script>
}

templ StatCard(title, value, textColor, bgColor string) {
    <div class="bg-white overflow-hidden shadow rounded-lg">
        <div class="px-4 py-5 sm:p-6">
            <dt class="text-sm font-medium text-gray-500 truncate">
                { title }
            </dt>
            <dd class={ "mt-1 text-3xl font-semibold", textColor }>
                <span class={ "text-sm px-2 py-1 rounded", bgColor }>
                    { value }
                </span>
            </dd>
        </div>
    </div>
}
```

### 3.3 Create Guest List View
Create `internal/views/admin_guest_list.templ`:

```templ
package views

import (
    "fmt"
    "github.com/apkiernan/thedrewzers/internal/models"
)

templ AdminGuestList(userName string, guests []*models.GuestWithRSVP) {
    @AdminLayout(userName, "Guests", AdminGuestListContent(guests))
}

templ AdminGuestListContent(guests []*models.GuestWithRSVP) {
    <div class="px-4 sm:px-0">
        <div class="sm:flex sm:items-center sm:justify-between mb-6">
            <h2 class="text-2xl font-bold text-gray-900">Guest List</h2>
            <div class="mt-4 sm:mt-0">
                <a 
                    href="/guests/add" 
                    class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
                >
                    Add Guest
                </a>
            </div>
        </div>
        
        // Search/Filter Bar
        <div class="mb-6 bg-white p-4 rounded-lg shadow">
            <input 
                type="text" 
                id="guestSearch"
                placeholder="Search guests..."
                class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500"
                onkeyup="filterGuests()"
            />
        </div>
        
        // Guest Table
        <div class="bg-white shadow rounded-lg overflow-hidden">
            <table class="min-w-full divide-y divide-gray-200" id="guestTable">
                <thead class="bg-gray-50">
                    <tr>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Guest</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Invitation Code</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Max Party</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">RSVP Status</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Response</th>
                        <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                    </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                    @for _, gwRSVP := range guests {
                        <tr class="guest-row">
                            <td class="px-6 py-4 whitespace-nowrap">
                                <div>
                                    <div class="text-sm font-medium text-gray-900 guest-name">
                                        { gwRSVP.Guest.PrimaryGuest }
                                    </div>
                                    if gwRSVP.Guest.Email != "" {
                                        <div class="text-sm text-gray-500">
                                            { gwRSVP.Guest.Email }
                                        </div>
                                    }
                                </div>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap">
                                <code class="text-sm bg-gray-100 px-2 py-1 rounded">
                                    { gwRSVP.Guest.InvitationCode }
                                </code>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                { fmt.Sprintf("%d", gwRSVP.Guest.MaxPartySize) }
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap">
                                if gwRSVP.RSVP != nil {
                                    <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                                        Responded
                                    </span>
                                } else {
                                    <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-yellow-100 text-yellow-800">
                                        Pending
                                    </span>
                                }
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                if gwRSVP.RSVP != nil {
                                    if gwRSVP.RSVP.Attending {
                                        { fmt.Sprintf("Yes (%d)", gwRSVP.RSVP.PartySize) }
                                    } else {
                                        No
                                    }
                                } else {
                                    -
                                }
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                <a href={ templ.SafeURL(fmt.Sprintf("/guests/edit?id=%s", gwRSVP.Guest.GuestID)) } class="text-indigo-600 hover:text-indigo-900">
                                    Edit
                                </a>
                            </td>
                        </tr>
                    }
                </tbody>
            </table>
        </div>
        
        // Summary
        <div class="mt-4 text-sm text-gray-500">
            Total: { fmt.Sprintf("%d guests", len(guests)) }
        </div>
    </div>
    
    <script>
        function filterGuests() {
            const input = document.getElementById('guestSearch');
            const filter = input.value.toUpperCase();
            const rows = document.querySelectorAll('.guest-row');
            
            rows.forEach(row => {
                const name = row.querySelector('.guest-name').textContent.toUpperCase();
                if (name.includes(filter)) {
                    row.style.display = '';
                } else {
                    row.style.display = 'none';
                }
            });
        }
    </script>
}
```

## Step 4: Testing the Dashboard

### 4.1 Create Test Data Generator
Create `cmd/generate-test-data/main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/rand"
    "time"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/google/uuid"
    
    dbRepo "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
    "github.com/apkiernan/thedrewzers/internal/models"
)

func main() {
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Fatal(err)
    }
    
    client := dynamodb.NewFromConfig(cfg)
    guestRepo := dbRepo.NewGuestRepository(client, "wedding-guests")
    rsvpRepo := dbRepo.NewRSVPRepository(client, "wedding-rsvps")
    
    // Create test guests
    testGuests := []models.Guest{
        {
            PrimaryGuest:   "Test Family 1",
            MaxPartySize:   4,
            InvitationCode: generateCode(),
            Email:         "test1@example.com",
        },
        {
            PrimaryGuest:   "Test Couple",
            MaxPartySize:   2,
            InvitationCode: generateCode(),
            Email:         "test2@example.com",
        },
        // Add more test guests
    }
    
    for _, guest := range testGuests {
        err := guestRepo.CreateGuest(context.TODO(), &guest)
        if err != nil {
            log.Printf("Failed to create guest: %v", err)
            continue
        }
        
        // Randomly create RSVPs
        if rand.Float32() < 0.7 { // 70% response rate
            rsvp := &models.RSVP{
                RSVPID:    uuid.New().String(),
                GuestID:   guest.GuestID,
                Attending: rand.Float32() < 0.85, // 85% attendance rate
                PartySize: rand.Intn(guest.MaxPartySize) + 1,
                SubmittedAt: time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
            }
            
            if rsvp.Attending {
                rsvp.DietaryRestrictions = []string{"Vegetarian", "Gluten-free"}[rand.Intn(2):]
            }
            
            err = rsvpRepo.CreateRSVP(context.TODO(), rsvp)
            if err != nil {
                log.Printf("Failed to create RSVP: %v", err)
            }
        }
    }
    
    fmt.Println("Test data generated successfully")
}

func generateCode() string {
    const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    b := make([]byte, 8)
    for i := range b {
        b[i] = charset[rand.Intn(len(charset))]
    }
    return string(b)
}
```

### 4.2 Test Admin Dashboard
```bash
# Login to admin
open https://admin.thedrewzers.com

# Test dashboard features:
# - View statistics
# - Check recent RSVPs
# - Export guest list
# - Search/filter guests
```

## Next Steps
- Phase 6: Testing and deployment
- Add guest editing functionality
- Implement email notifications
- Add more detailed reporting

## Features Checklist
- [x] Dashboard statistics
- [x] Guest list with search
- [x] RSVP list
- [x] CSV export
- [x] Recent activity
- [ ] Guest editing
- [ ] Bulk operations
- [ ] Email notifications
- [ ] Advanced filtering