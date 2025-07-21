package rng

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// Client for RNG service
type Client struct {
	ServiceURL string
}

// NewClient creates a new RNG client
func NewClient(serviceURL string) *Client {
	return &Client{
		ServiceURL: serviceURL,
	}
}

type Request struct {
	ClientID         string  `json:"client_id"`
	GameID           string  `json:"game_id"`
	PlayerID         string  `json:"player_id"`
	BetID            string  `json:"bet_id"`
	RTP              float64 `json:"rtp"`
	PayoutMultiplier float64 `json:"payout_multiplier"`
	RequestSalt      string  `json:"request_salt"`
	BetAmount        float64 `json:"bet_amount"`
	IPAddress        string  `json:"ip_address"`
    UserAgent        string  `json:"user_agent"`
}

type Response struct {
	PrefOutcome string  `json:"pref_outcome"`
	WinAmount   float64 `json:"win_amount"`
	WinProb     float64 `json:"win_prob"`
}

// GetOutcome calls the RNG service and returns the outcome
func (c *Client) GetOutcome(clientID, gameID, playerID, betID string, rtp, payoutMultiplier, betAmount float64, ipAddress string, userAgent string) (Response, error) {
	reqBody, err := json.Marshal(Request{
		ClientID:         clientID,
		GameID:           gameID,
		BetID:            betID,
		PlayerID:         playerID,
		RTP:              rtp,
		PayoutMultiplier: payoutMultiplier,
		RequestSalt:      uuid.New().String(),
		BetAmount:        betAmount,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
	})
	if err != nil {
		log.Printf("Error marshaling RNG request: %v", err)
		return Response{}, err
	}

	log.Printf("RNG request: %s", string(reqBody))

	resp, err := http.Post(c.ServiceURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error calling RNG API: %v", err)
		return Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("RNG API returned non-200 status: %d", resp.StatusCode)
		return Response{}, errors.New("RNG API call failed")
	}

	var rngResp Response
	if err := json.NewDecoder(resp.Body).Decode(&rngResp); err != nil {
		log.Printf("Error decoding RNG response: %v", err)
		return Response{}, err
	}

	return rngResp, nil
}
