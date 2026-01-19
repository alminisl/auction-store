package email

import (
	"fmt"
	"log"
)

type EmailType string

const (
	EmailVerification  EmailType = "verification"
	EmailPasswordReset EmailType = "password_reset"
	EmailOutbid        EmailType = "outbid"
	EmailAuctionWon    EmailType = "auction_won"
	EmailAuctionLost   EmailType = "auction_lost"
	EmailAuctionEnding EmailType = "auction_ending"
	EmailNewBid        EmailType = "new_bid"
)

type EmailData struct {
	To          string
	Subject     string
	Body        string
	Type        EmailType
	TemplateData map[string]interface{}
}

type Sender interface {
	Send(data *EmailData) error
}

// MockSender logs emails to console (for development)
type MockSender struct{}

func NewMockSender() *MockSender {
	return &MockSender{}
}

func (s *MockSender) Send(data *EmailData) error {
	log.Printf(`
========================================
EMAIL NOTIFICATION (Mock)
========================================
To: %s
Subject: %s
Type: %s
Body:
%s
========================================
`, data.To, data.Subject, data.Type, data.Body)
	return nil
}

// Helper functions to create common emails
func NewVerificationEmail(to, token, baseURL string) *EmailData {
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", baseURL, token)
	return &EmailData{
		To:      to,
		Subject: "Verify your email address",
		Type:    EmailVerification,
		Body: fmt.Sprintf(`
Welcome to Auction Marketplace!

Please verify your email address by clicking the link below:

%s

This link will expire in 24 hours.

If you did not create an account, please ignore this email.
`, verifyURL),
	}
}

func NewPasswordResetEmail(to, token, baseURL string) *EmailData {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)
	return &EmailData{
		To:      to,
		Subject: "Reset your password",
		Type:    EmailPasswordReset,
		Body: fmt.Sprintf(`
You requested to reset your password.

Click the link below to reset your password:

%s

This link will expire in 1 hour.

If you did not request a password reset, please ignore this email.
`, resetURL),
	}
}

func NewOutbidEmail(to, auctionTitle, newBidAmount, auctionURL string) *EmailData {
	return &EmailData{
		To:      to,
		Subject: fmt.Sprintf("You've been outbid on %s", auctionTitle),
		Type:    EmailOutbid,
		Body: fmt.Sprintf(`
You've been outbid!

Item: %s
New highest bid: %s

Don't miss out! Place a higher bid now:
%s
`, auctionTitle, newBidAmount, auctionURL),
	}
}

func NewAuctionWonEmail(to, auctionTitle, winningBid, auctionURL string) *EmailData {
	return &EmailData{
		To:      to,
		Subject: fmt.Sprintf("Congratulations! You won %s", auctionTitle),
		Type:    EmailAuctionWon,
		Body: fmt.Sprintf(`
Congratulations! You won the auction!

Item: %s
Winning bid: %s

View your won auction:
%s

The seller will contact you shortly with payment and shipping details.
`, auctionTitle, winningBid, auctionURL),
	}
}

func NewAuctionLostEmail(to, auctionTitle, winningBid, auctionURL string) *EmailData {
	return &EmailData{
		To:      to,
		Subject: fmt.Sprintf("Auction ended: %s", auctionTitle),
		Type:    EmailAuctionLost,
		Body: fmt.Sprintf(`
The auction has ended.

Item: %s
Winning bid: %s

Unfortunately, you didn't win this auction. Check out similar items:
%s
`, auctionTitle, winningBid, auctionURL),
	}
}

func NewAuctionEndingEmail(to, auctionTitle, timeRemaining, currentBid, auctionURL string) *EmailData {
	return &EmailData{
		To:      to,
		Subject: fmt.Sprintf("Auction ending soon: %s", auctionTitle),
		Type:    EmailAuctionEnding,
		Body: fmt.Sprintf(`
An auction you're watching is ending soon!

Item: %s
Time remaining: %s
Current bid: %s

Don't miss out! Place your bid now:
%s
`, auctionTitle, timeRemaining, currentBid, auctionURL),
	}
}

func NewNewBidEmail(to, auctionTitle, bidAmount, bidderName, auctionURL string) *EmailData {
	return &EmailData{
		To:      to,
		Subject: fmt.Sprintf("New bid on your auction: %s", auctionTitle),
		Type:    EmailNewBid,
		Body: fmt.Sprintf(`
You received a new bid!

Item: %s
Bid amount: %s
Bidder: %s

View your auction:
%s
`, auctionTitle, bidAmount, bidderName, auctionURL),
	}
}
