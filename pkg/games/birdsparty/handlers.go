package birdsparty

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SpinHandler handles the /spin/birdsparty endpoint
// Generates grid with potential wins, identifies stage-cleared symbols
// Checks for regular bird symbol connections to determine cascading
func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
	rngClient, settingsClient := rg.getClientsForRequest(c)
	
	var req SpinRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.GameState.Bet.Amount); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Initialize game state if needed
	if req.GameState.CurrentLevel == 0 {
		req.GameState = InitializeGameState()
	}
	if req.GameState.GameMode == "" {
		req.GameState.GameMode = "base"
	}

	// Ensure grid size matches current level
	expectedGridSize := req.GameState.CurrentLevel.GetGridSize()
	if req.GameState.GridSize != expectedGridSize {
		req.GameState.GridSize = expectedGridSize
		log.Printf("Corrected grid size to %d for level %d", expectedGridSize, req.GameState.CurrentLevel)
	}

	// Create rand instance
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Set bet multiplier
	req.GameState.Bet.Multiplier = BetAmountToMultiplier[req.GameState.Bet.Amount]

	// Generate grid with potential bird symbol connections
	req.GameState.Grid = GenerateGridWithWin(req.GameState.CurrentLevel, r)

	// Find stage-cleared symbols (do NOT remove them yet)
	stageClearedSymbols := FindStageClearedSymbols(req.GameState.Grid, req.GameState.CurrentLevel)
	req.GameState.StageClearedSymbols = stageClearedSymbols

	// Check for regular bird symbol connections to determine if cascading will happen
	connections := FindRegularConnections(req.GameState.Grid, req.GameState.CurrentLevel)
	
	// Calculate total winnings
	totalWinnings := 0.0
	for i, connection := range connections {
		multiplier := 1.0
		if req.GameState.GameMode == "freeSpins" {
			multiplier = req.GameState.FreeSpins.Multiplier
		}
		connections[i].Payout = calculatePayout(connection.Symbol, connection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		connections[i].Payout *= multiplier
		totalWinnings += connections[i].Payout
	}

	// Get RTP and call RNG for bird symbol connections
	if len(connections) > 0 {
		rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
		if err != nil {
			log.Printf("Failed to get RTP: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve game settings",
			})
		}

		// Call RNG
		payoutMultiplier := totalWinnings / req.GameState.Bet.Amount
		rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.GameState.Bet.Amount)
		if err != nil {
			log.Printf("Failed to call RNG API: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to determine outcome",
			})
		}

		// Adjust outcome based on RNG
		if rngResp.PrefOutcome == "loss" {
			log.Printf("RNG determined a loss outcome")
			req.GameState.Grid = GenerateLossGrid(req.GameState.CurrentLevel, r)
			
			// Re-find stage-cleared symbols in loss grid
			stageClearedSymbols = FindStageClearedSymbols(req.GameState.Grid, req.GameState.CurrentLevel)
			req.GameState.StageClearedSymbols = stageClearedSymbols
			
			connections = nil
			totalWinnings = 0
		}
	}

	// Reset cascade count for new spin
	req.GameState.CascadeCount = 0
	req.GameState.TotalWin = totalWinnings
	req.GameState.LastConnections = connections
	req.GameState.Cascading = len(connections) > 0

	// Check for free game symbols
	freeGameCount := CountFreeGameSymbols(req.GameState.Grid)
	if req.GameState.GameMode == "base" && freeGameCount > 0 {
		req.GameState.GameMode = "freeSpins"
		req.GameState.FreeSpins.Remaining = FreeSpinsAwarded
		req.GameState.FreeSpins.TotalAwarded = FreeSpinsAwarded
		req.GameState.FreeSpins.Multiplier = GetRandomFreeSpinMultiplier(r)
		log.Printf("Free Spins triggered with %.1fx multiplier", req.GameState.FreeSpins.Multiplier)
	}

	// Update free spins count
	if req.GameState.GameMode == "freeSpins" {
		req.GameState.FreeSpins.Remaining--
		if req.GameState.FreeSpins.Remaining <= 0 {
			req.GameState.GameMode = "base"
			req.GameState.FreeSpins = struct {
				Remaining    int     `json:"remaining"`
				TotalAwarded int     `json:"totalAwarded"`
				Multiplier   float64 `json:"multiplier"`
			}{0, 0, 1.0}
			log.Printf("Free Spins ended")
		}
	}

	// Calculate total cost
	var totalCost float64
	if req.GameState.GameMode == "freeSpins" {
		totalCost = 0
	} else {
		totalCost = req.GameState.Bet.Amount
	}

	// Determine if we have stage-cleared symbols
	hasStageCleared := len(stageClearedSymbols) > 0

	log.Printf("Spin completed: level=%d, gridSize=%dx%d, stageClearedSymbols=%d, hasStageCleared=%v, cascading=%v", 
		req.GameState.CurrentLevel, req.GameState.GridSize, req.GameState.GridSize, 
		len(stageClearedSymbols), hasStageCleared, req.GameState.Cascading)

	return c.JSON(SpinResponse{
		Status:              "success",
		Message:             "",
		GameState:           req.GameState,
		StageClearedSymbols: stageClearedSymbols,
		HasStageCleared:     hasStageCleared,
		TotalCost:           totalCost,
	})
}

// ProcessStageClearedHandler handles the /process-stage-cleared/birdsparty endpoint
// This endpoint removes stage-cleared symbols, applies gravity, checks for level advancement
// AND checks for regular bird symbol connections in the new grid
func (rg *RouteGroup) ProcessStageClearedHandler(c *fiber.Ctx) error {
	rngClient, settingsClient := rg.getClientsForRequest(c)
	
	var req ProcessStageClearedRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.GameState.Bet.Amount); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate grid dimensions
	if !ValidateGridDimensions(req.GameState.Grid, req.GameState.CurrentLevel) {
		log.Printf("Invalid grid dimensions for level %d", req.GameState.CurrentLevel)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid grid dimensions",
		})
	}

	// Create rand instance
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get stage-cleared symbols from the current grid
	stageClearedSymbols := req.GameState.StageClearedSymbols
	if len(stageClearedSymbols) == 0 {
		// If none provided, find them from the grid
		stageClearedSymbols = FindStageClearedSymbols(req.GameState.Grid, req.GameState.CurrentLevel)
	}

	stageClearedCount := len(stageClearedSymbols)
	
	// Process stage-cleared symbols (remove, apply gravity, check level advancement)
	levelAdvanced, oldLevel, newLevel := ProcessStageClearedSymbols(
		&req.GameState, stageClearedSymbols, req.GameState.CurrentLevel, r)

	// Clear the stage-cleared symbols from game state since they've been processed
	req.GameState.StageClearedSymbols = []StageClearedSymbol{}

	// NOW check for regular bird symbol connections in the new grid after gravity
	connections := FindRegularConnections(req.GameState.Grid, req.GameState.CurrentLevel)
	
	// Calculate total winnings
	totalWinnings := 0.0
	for i, connection := range connections {
		multiplier := 1.0
		if req.GameState.GameMode == "freeSpins" {
			multiplier = req.GameState.FreeSpins.Multiplier
		}
		connections[i].Payout = calculatePayout(connection.Symbol, connection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		connections[i].Payout *= multiplier
		totalWinnings += connections[i].Payout
	}

	// Get RTP and call RNG for bird symbol connections (if any)
	if len(connections) > 0 {
		rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
		if err != nil {
			log.Printf("Failed to get RTP: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve game settings",
			})
		}

		// Call RNG
		payoutMultiplier := totalWinnings / req.GameState.Bet.Amount
		rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.GameState.Bet.Amount)
		if err != nil {
			log.Printf("Failed to call RNG API: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to determine outcome",
			})
		}

		// Adjust outcome based on RNG
		if rngResp.PrefOutcome == "loss" {
			log.Printf("RNG determined a loss outcome for connections after stage-cleared processing")
			// Force loss grid but maintain level advancement
			req.GameState.Grid = ForceLossGrid(req.GameState.CurrentLevel, r)
			connections = nil
			totalWinnings = 0
		}
	}

	// Update game state with connection results
	req.GameState.TotalWin = totalWinnings
	req.GameState.LastConnections = connections
	req.GameState.Cascading = len(connections) > 0

	// Reset cascade count since this is after stage-cleared processing
	req.GameState.CascadeCount = 0

	log.Printf("ProcessStageCleared completed: stageClearedCount=%d, levelAdvanced=%v, oldLevel=%d, newLevel=%d, progress=%d, cascading=%v", 
		stageClearedCount, levelAdvanced, oldLevel, newLevel, req.GameState.StageProgress, req.GameState.Cascading)

	return c.JSON(ProcessStageClearedResponse{
		Status:            "success",
		Message:           "",
		GameState:         req.GameState,
		StageClearedCount: stageClearedCount,
		LevelAdvanced:     levelAdvanced,
		OldLevel:          oldLevel,
		NewLevel:          newLevel,
		Connections:       connections,
		TotalCost:         0,
	})
}

// CascadeHandler handles the /cascade/birdsparty endpoint
// This endpoint processes regular bird symbol connections and DETECTS stage-cleared symbols
// It does NOT process stage-cleared symbols - client must call process-stage-cleared endpoint
func (rg *RouteGroup) CascadeHandler(c *fiber.Ctx) error {
	rngClient, settingsClient := rg.getClientsForRequest(c)
	
	var req CascadeRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.GameState.Bet.Amount); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate grid dimensions
	if !ValidateGridDimensions(req.GameState.Grid, req.GameState.CurrentLevel) {
		log.Printf("Invalid grid dimensions for level %d", req.GameState.CurrentLevel)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid grid dimensions",
		})
	}

	// Increment cascade count
	req.GameState.CascadeCount++

	// Create rand instance
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// If this is NOT the first cascade call, remove previous connections and apply gravity first
	var connections []Connection
	var totalWinnings float64

	if req.GameState.CascadeCount > 1 && len(req.GameState.LastConnections) > 0 {
		// Remove previous connections and apply gravity
		RemoveConnections(req.GameState.Grid, req.GameState.LastConnections)
		ApplyGravity(req.GameState.Grid, req.GameState.CurrentLevel, r)
	}
	
	// Find regular bird symbol connections
	connections = FindRegularConnections(req.GameState.Grid, req.GameState.CurrentLevel)

	// Calculate total winnings
	for i, connection := range connections {
		multiplier := 1.0
		if req.GameState.GameMode == "freeSpins" {
			multiplier = req.GameState.FreeSpins.Multiplier
		}
		connections[i].Payout = calculatePayout(connection.Symbol, connection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		connections[i].Payout *= multiplier
		totalWinnings += connections[i].Payout
	}

	// Get RTP and call RNG only if there are bird connections
	if len(connections) > 0 {
		rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
		if err != nil {
			log.Printf("Failed to get RTP: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve game settings",
			})
		}

		// Call RNG
		payoutMultiplier := totalWinnings / req.GameState.Bet.Amount
		rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.GameState.Bet.Amount)
		if err != nil {
			log.Printf("Failed to call RNG API: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to determine outcome",
			})
		}

		// Adjust outcome based on RNG
		if rngResp.PrefOutcome == "loss" {
			log.Printf("RNG determined a loss outcome")
			// Generate grid without connections
			req.GameState.Grid = ForceLossGrid(req.GameState.CurrentLevel, r)
			connections = nil
			totalWinnings = 0
		}
	}

	// IMPORTANT: After all processing, check for stage-cleared symbols that may have appeared
	stageClearedSymbols := FindStageClearedSymbols(req.GameState.Grid, req.GameState.CurrentLevel)
	hasStageCleared := len(stageClearedSymbols) > 0
	
	// Store stage-cleared symbols in game state for potential next call to process-stage-cleared
	req.GameState.StageClearedSymbols = stageClearedSymbols

	// Check for free game symbols if new grid was generated
	if len(connections) == 0 {
		freeGameCount := CountFreeGameSymbols(req.GameState.Grid)
		if req.GameState.GameMode == "base" && freeGameCount > 0 {
			req.GameState.GameMode = "freeSpins"
			req.GameState.FreeSpins.Remaining = FreeSpinsAwarded
			req.GameState.FreeSpins.TotalAwarded = FreeSpinsAwarded
			req.GameState.FreeSpins.Multiplier = GetRandomFreeSpinMultiplier(r)
			log.Printf("Free Spins triggered during cascade with %.1fx multiplier", req.GameState.FreeSpins.Multiplier)
		}
	}

	// Update game state
	req.GameState.TotalWin = totalWinnings
	req.GameState.LastConnections = connections
	req.GameState.Cascading = len(connections) > 0

	log.Printf("Cascade completed: level=%d, gridSize=%dx%d, totalWin=%.2f, cascading=%v, cascadeCount=%d, stageClearedDetected=%v", 
		req.GameState.CurrentLevel, req.GameState.GridSize, req.GameState.GridSize, 
		totalWinnings, req.GameState.Cascading, req.GameState.CascadeCount, hasStageCleared)

	return c.JSON(CascadeResponse{
		Status:              "success",
		Message:             "",
		GameState:           req.GameState,
		Connections:         connections,
		StageClearedSymbols: stageClearedSymbols,  // NEW: Include detected stage-cleared symbols
		HasStageCleared:     hasStageCleared,      // NEW: Flag to indicate stage-cleared symbols found
		TotalCost:           0,
	})
}

// validateRequest validates the request fields
func validateRequest(clientID, gameID, playerID, betID string, betAmount float64) error {
	if clientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if gameID == "" {
		return fmt.Errorf("game_id is required")
	}
	if playerID == "" {
		return fmt.Errorf("player_id is required")
	}
	if betID == "" {
		return fmt.Errorf("bet_id is required")
	}
	if !isValidBetAmount(betAmount) {
		return fmt.Errorf("invalid bet amount, allowed values are 0.1, 0.2, 0.3, 0.5, 1.0")
	}
	return nil
}

// isValidBetAmount checks if the bet amount is valid
func isValidBetAmount(amount float64) bool {
	validAmounts := []float64{0.1, 0.2, 0.3, 0.5, 1.0}
	for _, valid := range validAmounts {
		if amount == valid {
			return true
		}
	}
	return false
}