package qrcode

import (
	"fmt"
	"image/color"

	qr "github.com/skip2/go-qrcode"
)

// Generator creates QR codes for wedding invitations
type Generator struct {
	baseURL string
}

// NewGenerator creates a QR code generator with the given base URL
func NewGenerator(baseURL string) *Generator {
	return &Generator{baseURL: baseURL}
}

// GenerateInvitationQR creates a QR code for a specific invitation code
// Returns the PNG image data as bytes
func (g *Generator) GenerateInvitationQR(invitationCode string) ([]byte, error) {
	rsvpURL := fmt.Sprintf("%s/rsvp?code=%s", g.baseURL, invitationCode)

	qrCode, err := qr.New(rsvpURL, qr.High)
	if err != nil {
		return nil, fmt.Errorf("creating QR code: %w", err)
	}

	qrCode.BackgroundColor = color.White
	qrCode.ForegroundColor = color.Black

	png, err := qrCode.PNG(512)
	if err != nil {
		return nil, fmt.Errorf("generating PNG: %w", err)
	}

	return png, nil
}

// Style configures QR code appearance
type Style struct {
	Size            int
	BackgroundColor color.Color
	ForegroundColor color.Color
}

// DefaultStyle returns the default QR code style (black on white, 512px)
func DefaultStyle() Style {
	return Style{
		Size:            512,
		BackgroundColor: color.White,
		ForegroundColor: color.Black,
	}
}

// WeddingStyle returns a wedding-themed QR style
// Uses a softer dark color that prints well while looking elegant
func WeddingStyle() Style {
	return Style{
		Size:            512,
		BackgroundColor: color.White,
		ForegroundColor: color.RGBA{R: 64, G: 64, B: 64, A: 255}, // Dark gray
	}
}

// GenerateStyledQR creates a QR code with custom styling
func (g *Generator) GenerateStyledQR(invitationCode string, style Style) ([]byte, error) {
	rsvpURL := fmt.Sprintf("%s/rsvp?code=%s", g.baseURL, invitationCode)

	qrCode, err := qr.New(rsvpURL, qr.High)
	if err != nil {
		return nil, fmt.Errorf("creating QR code: %w", err)
	}

	qrCode.BackgroundColor = style.BackgroundColor
	qrCode.ForegroundColor = style.ForegroundColor

	png, err := qrCode.PNG(style.Size)
	if err != nil {
		return nil, fmt.Errorf("generating PNG: %w", err)
	}

	return png, nil
}

// GetRSVPURL returns the full RSVP URL for an invitation code
// Useful for testing or generating links without QR codes
func (g *Generator) GetRSVPURL(invitationCode string) string {
	return fmt.Sprintf("%s/rsvp?code=%s", g.baseURL, invitationCode)
}
