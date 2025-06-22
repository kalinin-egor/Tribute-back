package payouts

import (
	"fmt"
	"log"
)

// CardDetails holds the necessary (but sensitive) card information.
// In a real application, this should never be stored.
type CardDetails struct {
	CardNumber string
	CardDate   string
	CardCVV    string
}

// Gateway defines the interface for a payout provider.
type Gateway interface {
	RegisterPayoutMethod(userID int64, details CardDetails) error
}

// MockGateway is a simulated implementation of a payment gateway.
type MockGateway struct{}

// NewMockGateway creates a new mock gateway.
func NewMockGateway() Gateway {
	return &MockGateway{}
}

// RegisterPayoutMethod simulates registering a user's card with a third-party service.
// It does NOT store the card details.
func (g *MockGateway) RegisterPayoutMethod(userID int64, details CardDetails) error {
	// In a real implementation, you would make an API call to your payment provider here.
	// For example, with Stripe, you would create a token and then a customer or payout destination.
	log.Printf("Simulating payout method registration for user %d with card ending in %s",
		userID, details.CardNumber[len(details.CardNumber)-4:])

	// Simulate a successful API call
	if details.CardCVV == "123" { // Simulate a failure for a specific CVV
		return fmt.Errorf("mock gateway error: invalid card details")
	}

	return nil
}
