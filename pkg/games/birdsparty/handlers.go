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
	forbidFreeGame := req.GameState.GameMode == "freeSpins"
	req.GameState.Grid = GenerateGridWithWin(req.GameState.CurrentLevel, r, forbidFreeGame)

	// Find stage-cleared symbols (do NOT remove them yet)
	stageClearedSymbols := FindStageClearedSymbols(req.GameState.Grid, req.GameState.CurrentLevel)
	req.GameState.StageClearedSymbols = stageClearedSymbols

	// Check for regular bird symbol connections to determine if cascading will happen
	connections := FindRegularConnections(req.GameState.Grid, req.GameState.CurrentLevel)

	// Calculate total winnings
	totalWinnings := 0.0
	for i, connection := range connections {
		payout := calculatePayout(connection.Symbol, connection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		connections[i].Payout = payout // base payout, no multiplier
		totalWinnings += payout
	}
	if req.GameState.GameMode == "freeSpins" {
		totalWinnings *= req.GameState.FreeSpins.Multiplier
	}
	totalWinnings = round(totalWinnings)

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
			forbidFreeGame := req.GameState.GameMode == "freeSpins"
			req.GameState.Grid = GenerateLossGrid(req.GameState.CurrentLevel, r, forbidFreeGame)

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
// FIXED: Now preserves grid structure and uses surgical loss approach
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

	// PRESERVE ORIGINAL GRID before processing
	originalGrid := make([][]string, len(req.GameState.Grid))
	for i := range req.GameState.Grid {
		originalGrid[i] = make([]string, len(req.GameState.Grid[i]))
		copy(originalGrid[i], req.GameState.Grid[i])
	}

	// Prepare variables for level advancement
	var (
		levelAdvanced bool
		oldLevel      = req.GameState.CurrentLevel
		newLevel      = req.GameState.CurrentLevel
	)

	// Process stage-cleared symbols (remove, apply gravity, check level advancement)
	var newPositions []Position
	{
		// Remove stage-cleared symbols from grid surgically
		RemoveStageClearedSymbolsSurgical(req.GameState.Grid, stageClearedSymbols)
		// Apply gravity surgically and get new positions
		newPositions = ApplyGravitySurgical(req.GameState.Grid, stageClearedSymbols, req.GameState.CurrentLevel, r, req.GameState.GameMode == "freeSpins")
		// Update stage progress
		req.GameState.StageProgress += len(stageClearedSymbols)
		log.Printf("Added %d stage-cleared symbols to progress, total: %d/15", len(stageClearedSymbols), req.GameState.StageProgress)
		// Check for level advancement
		if req.GameState.StageProgress >= StageProgressTarget {
			newLevel = AdvanceLevel(oldLevel)
			excessProgress := req.GameState.StageProgress - StageProgressTarget
			UpdateGameStateForLevel(&req.GameState, newLevel)
			req.GameState.StageProgress = excessProgress
			req.GameState.Grid = GenerateGrid(newLevel, r, false) // No free game in new level
			levelAdvanced = true
			log.Printf("Level advanced from %d to %d, excess progress: %d", oldLevel, newLevel, excessProgress)
			// Clear stage-cleared symbols
			req.GameState.StageClearedSymbols = []StageClearedSymbol{}
			return c.JSON(ProcessStageClearedResponse{
				Status:            "success",
				Message:           "",
				GameState:         req.GameState,
				StageClearedCount: len(stageClearedSymbols),
				LevelAdvanced:     levelAdvanced,
				OldLevel:          oldLevel,
				NewLevel:          newLevel,
				Connections:       nil,
				TotalCost:         0,
			})
		}
	}
	// Clear the stage-cleared symbols from game state since they've been processed
	req.GameState.StageClearedSymbols = []StageClearedSymbol{}

	// NOW check for regular bird symbol connections in the new grid after gravity
	connections := FindRegularConnections(req.GameState.Grid, req.GameState.CurrentLevel)

	// Calculate total winnings
	totalWinnings := 0.0
	for i, connection := range connections {
		payout := calculatePayout(connection.Symbol, connection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		connections[i].Payout = payout // base payout, no multiplier
		totalWinnings += payout
	}
	if req.GameState.GameMode == "freeSpins" {
		totalWinnings *= req.GameState.FreeSpins.Multiplier
	}
	totalWinnings = round(totalWinnings)

	// Handle RNG for bird symbol connections (if any) with surgical loss approach
	rngBypassed := false
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

		// SURGICAL LOSS: Adjust outcome based on RNG while preserving grid structure
		if rngResp.PrefOutcome == "loss" {
			log.Printf("RNG determined a loss outcome for stage-cleared processing")
			// Try surgical loss approach first (only new positions)
			success := ApplySurgicalLoss(&req.GameState, originalGrid, stageClearedSymbols, req.GameState.CurrentLevel, r, newPositions)
			if !success {
				// If surgical loss is impossible, bypass RNG and allow the win
				log.Printf("⚠️  RNG BYPASS: Surgical loss impossible after stage-cleared processing - preserving natural outcome")
				log.Printf("⚠️  GRID PRESERVATION: Maintaining grid structure as surgical loss would break game mechanics")
				log.Printf("⚠️  REASON: Stage-cleared symbol removal at positions %+v made loss impossible", stageClearedSymbols)
				rngBypassed = true
				// Keep the original connections and winnings
				// Grid remains as it is after stage-cleared processing
			} else {
				// Surgical loss successful - remove connections
				connections = nil
				totalWinnings = 0
				log.Printf("Surgical loss applied successfully after stage-cleared processing (new only)")
			}
		}
	}

	// Update game state with connection results
	req.GameState.TotalWin = totalWinnings
	req.GameState.LastConnections = connections
	req.GameState.Cascading = len(connections) > 0

	// Reset cascade count since this is after stage-cleared processing
	// if there is a win then cascade count to be 1 else 0
	if len(connections) > 0 {
		req.GameState.CascadeCount = 1
	} else {
		req.GameState.CascadeCount = 0
	}

	logMessage := fmt.Sprintf("ProcessStageCleared completed: stageClearedCount=%d, levelAdvanced=%v, oldLevel=%d, newLevel=%d, progress=%d, cascading=%v",
		stageClearedCount, levelAdvanced, oldLevel, newLevel, req.GameState.StageProgress, req.GameState.Cascading)

	if rngBypassed {
		logMessage += " [RNG BYPASSED - Surgical loss impossible]"
	}

	log.Printf(logMessage)

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
// FIXED: Now uses surgical cascade processing to preserve grid structure
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

	// PRESERVE ORIGINAL GRID before processing for surgical loss capability
	originalGrid := make([][]string, len(req.GameState.Grid))
	for i := range req.GameState.Grid {
		originalGrid[i] = make([]string, len(req.GameState.Grid[i]))
		copy(originalGrid[i], req.GameState.Grid[i])
	}

	var connections []Connection
	var totalWinnings float64
	var affectedPositions []Position
	var newPositions []Position

	// Process cascade surgically
	if req.GameState.CascadeCount >= 1 && len(req.GameState.LastConnections) > 0 {
		// SURGICAL: Remove previous connections and apply gravity surgically
		affectedPositions = RemoveConnectionsSurgical(req.GameState.Grid, req.GameState.LastConnections)
		newPositions = ApplyGravitySurgicalForCascade(req.GameState.Grid, affectedPositions, req.GameState.CurrentLevel, r, req.GameState.GameMode == "freeSpins")
	} else {
		// First cascade call - find existing connections
		connections = FindRegularConnections(req.GameState.Grid, req.GameState.CurrentLevel)
		if len(connections) > 0 {
			// Extract positions that will be affected for surgical processing
			for _, connection := range connections {
				affectedPositions = append(affectedPositions, connection.Positions...)
			}
			newPositions = ApplyGravitySurgicalForCascade(req.GameState.Grid, affectedPositions, req.GameState.CurrentLevel, r, req.GameState.GameMode == "freeSpins")
		}
	}

	// Find regular bird symbol connections after cascade processing
	if req.GameState.CascadeCount >= 1 || len(connections) == 0 {
		connections = FindRegularConnections(req.GameState.Grid, req.GameState.CurrentLevel)
	}

	// Calculate total winnings
	totalWinnings = 0.0
	for i, connection := range connections {
		payout := calculatePayout(connection.Symbol, connection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		connections[i].Payout = payout // base payout, no multiplier
		totalWinnings += payout
	}
	if req.GameState.GameMode == "freeSpins" {
		totalWinnings *= req.GameState.FreeSpins.Multiplier
	}
	totalWinnings = round(totalWinnings)

	// Handle RNG for bird symbol connections with surgical loss approach
	rngBypassed := false
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

		// SURGICAL LOSS: Adjust outcome based on RNG while preserving grid structure
		if rngResp.PrefOutcome == "loss" {
			log.Printf("RNG determined a loss outcome for cascade")

			// Try surgical loss approach first (only new positions)
			success := ApplySurgicalLossForCascade(&req.GameState, originalGrid, newPositions, req.GameState.CurrentLevel, r)

			if !success {
				// If surgical loss is impossible, bypass RNG and allow the win
				log.Printf("⚠️  RNG BYPASS: Surgical loss impossible after cascade processing - preserving natural outcome")
				log.Printf("⚠️  GRID PRESERVATION: Maintaining grid structure as surgical loss would break game mechanics")
				log.Printf("⚠️  REASON: Cascade processing at %d positions made loss impossible", len(affectedPositions))
				rngBypassed = true

				// Keep the original connections and winnings
				// Grid remains as it is after cascade processing
			} else {
				// Surgical loss successful - remove connections
				connections = nil
				totalWinnings = 0
				log.Printf("Surgical loss applied successfully after cascade processing (new only)")
			}
		}
	}

	// IMPORTANT: After all processing, check for stage-cleared symbols that may have appeared
	stageClearedSymbols := FindStageClearedSymbols(req.GameState.Grid, req.GameState.CurrentLevel)
	hasStageCleared := len(stageClearedSymbols) > 0

	// Store stage-cleared symbols in game state for potential next call to process-stage-cleared
	req.GameState.StageClearedSymbols = stageClearedSymbols

	// Check for free game symbols if connections were removed by RNG or no connections exist
	if len(connections) == 0 && !rngBypassed {
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

	logMessage := fmt.Sprintf("Cascade completed: level=%d, gridSize=%dx%d, totalWin=%.2f, cascading=%v, cascadeCount=%d, stageClearedDetected=%v",
		req.GameState.CurrentLevel, req.GameState.GridSize, req.GameState.GridSize,
		totalWinnings, req.GameState.Cascading, req.GameState.CascadeCount, hasStageCleared)

	if rngBypassed {
		logMessage += " [RNG BYPASSED - Surgical loss impossible]"
	}

	log.Printf(logMessage)

	return c.JSON(CascadeResponse{
		Status:              "success",
		Message:             "",
		GameState:           req.GameState,
		Connections:         connections,
		StageClearedSymbols: stageClearedSymbols, // Include detected stage-cleared symbols
		HasStageCleared:     hasStageCleared,     // Flag to indicate stage-cleared symbols found
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
