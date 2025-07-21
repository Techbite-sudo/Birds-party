package birdsparty

import "fmt"

// Symbol type
type Symbol string

const (
	// Regular bird symbols
	SymbolPurpleOwl Symbol = "purple_owl"
	SymbolGreenOwl  Symbol = "green_owl"
	SymbolYellowOwl Symbol = "yellow_owl"
	SymbolBlueOwl   Symbol = "blue_owl"
	SymbolRedOwl    Symbol = "red_owl"

	// Special symbols
	SymbolFreeGame Symbol = "free_game"

	// Stage-cleared symbols (level-specific)
	SymbolOrangeSlice Symbol = "orange_slice" // Level 1 stage-cleared symbol
	SymbolHoneyPot    Symbol = "honey_pot"    // Level 2 stage-cleared symbol
	SymbolStrawberry  Symbol = "strawberry"   // Level 3 stage-cleared symbol
)

// Game constants
const (
	MinBet = 10

	// Level requirements and grid sizes
	Level1MinConnection = 4
	Level2MinConnection = 5
	Level3MinConnection = 6

	Level1GridSize = 4 // 4x4 = 16 positions
	Level2GridSize = 5 // 5x5 = 25 positions
	Level3GridSize = 6 // 6x6 = 36 positions

	// Progress requirements
	StageProgressTarget = 15

	// Free spin settings
	FreeSpinsAwarded = 10
)

// Current level type
type Level int

const (
	Level1 Level = 1
	Level2 Level = 2
	Level3 Level = 3
)

// Position represents a position on the grid
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// StageClearedSymbol represents a stage-cleared symbol found on the grid
type StageClearedSymbol struct {
	Symbol   Symbol   `json:"symbol"`
	Position Position `json:"position"`
}

// Connection represents a group of connected symbols
type Connection struct {
	Symbol    Symbol     `json:"symbol"`
	Positions []Position `json:"positions"`
	Count     int        `json:"count"`
	Payout    float64    `json:"payout"`
}

// GameState represents the current state of the game
type GameState struct {
	Bet struct {
		Amount     float64 `json:"amount"`
		Multiplier int     `json:"multiplier"`
	} `json:"bet"`
	CurrentLevel  Level      `json:"currentLevel"`
	GridSize      int        `json:"gridSize"`      // Current grid dimensions (4, 5, or 6)
	Grid          [][]string `json:"grid"`          // Dynamic grid size
	StageProgress int        `json:"stageProgress"` // Accumulated stage-cleared symbols (0-14)
	GameMode      string     `json:"gameMode"`      // "base" or "freeSpins"
	FreeSpins     struct {
		Remaining    int     `json:"remaining"`
		TotalAwarded int     `json:"totalAwarded"`
		Multiplier   float64 `json:"multiplier"` // 1.0-5.0x random multiplier
	} `json:"freeSpins"`
	TotalWin        float64      `json:"totalWin"`
	Cascading       bool         `json:"cascading"`
	LastConnections []Connection `json:"lastConnections"`
	CascadeCount    int          `json:"cascadeCount"`
	// New field for tracking stage-cleared symbols in current spin
	StageClearedSymbols []StageClearedSymbol `json:"stageClearedSymbols"`
}

// SpinRequest represents the request body for the /spin endpoint
type SpinRequest struct {
	GameState GameState `json:"gameState"`
	ClientID  string    `json:"client_id"`
	GameID    string    `json:"game_id"`
	PlayerID  string    `json:"player_id"`
	BetID     string    `json:"bet_id"`
}

// ProcessStageClearedRequest represents the request body for the /process-stage-cleared endpoint
type ProcessStageClearedRequest struct {
	GameState GameState `json:"gameState"`
	ClientID  string    `json:"client_id"`
	GameID    string    `json:"game_id"`
	PlayerID  string    `json:"player_id"`
	BetID     string    `json:"bet_id"`
}

// CascadeRequest represents the request body for the /cascade endpoint
type CascadeRequest struct {
	GameState GameState `json:"gameState"`
	ClientID  string    `json:"client_id"`
	GameID    string    `json:"game_id"`
	PlayerID  string    `json:"player_id"`
	BetID     string    `json:"bet_id"`
}

// SpinResponse represents the response body for the /spin endpoint
type SpinResponse struct {
	Status              string               `json:"status"`
	Message             string               `json:"message"`
	GameState           GameState            `json:"gameState"`
	StageClearedSymbols []StageClearedSymbol `json:"stageClearedSymbols"`
	HasStageCleared     bool                 `json:"hasStageCleared"`
	TotalCost           float64              `json:"totalCost"`
}

// ProcessStageClearedResponse represents the response body for the /process-stage-cleared endpoint
type ProcessStageClearedResponse struct {
	Status            string       `json:"status"`
	Message           string       `json:"message"`
	GameState         GameState    `json:"gameState"`
	StageClearedCount int          `json:"stageClearedCount"`
	LevelAdvanced     bool         `json:"levelAdvanced"`
	OldLevel          Level        `json:"oldLevel,omitempty"`
	NewLevel          Level        `json:"newLevel,omitempty"`
	Connections       []Connection `json:"connections"`
	TotalCost         float64      `json:"totalCost"`
}

// CascadeResponse represents the response body for the /cascade endpoint
type CascadeResponse struct {
	Status              string               `json:"status"`
	Message             string               `json:"message"`
	GameState           GameState            `json:"gameState"`
	Connections         []Connection         `json:"connections"`
	StageClearedSymbols []StageClearedSymbol `json:"stageClearedSymbols"`
	HasStageCleared     bool                 `json:"hasStageCleared"`
	TotalCost           float64              `json:"totalCost"`
}

// ValidateLevel validates the current level
func (l Level) ValidateLevel() error {
	switch l {
	case Level1, Level2, Level3:
		return nil
	default:
		return fmt.Errorf("invalid level: %d", l)
	}
}

// GetMinConnection returns the minimum connection requirement for the level
func (l Level) GetMinConnection() int {
	switch l {
	case Level1:
		return Level1MinConnection
	case Level2:
		return Level2MinConnection
	case Level3:
		return Level3MinConnection
	default:
		return Level1MinConnection
	}
}

// GetGridSize returns the grid size for the level
func (l Level) GetGridSize() int {
	switch l {
	case Level1:
		return Level1GridSize
	case Level2:
		return Level2GridSize
	case Level3:
		return Level3GridSize
	default:
		return Level1GridSize
	}
}

// GetStageClearedSymbol returns the stage-cleared symbol for the level
func (l Level) GetStageClearedSymbol() Symbol {
	switch l {
	case Level1:
		return SymbolOrangeSlice
	case Level2:
		return SymbolHoneyPot
	case Level3:
		return SymbolStrawberry
	default:
		return SymbolOrangeSlice
	}
}

// BetAmountToMultiplier maps bet amounts to multipliers
var BetAmountToMultiplier = map[float64]int{
	0.1: 1,
	0.2: 2,
	0.3: 3,
	0.5: 5,
	1.0: 10,
	2.0: 20,
	2.5: 25,
}

// Symbol weights for random generation (global across all levels)
var SymbolWeights = map[Symbol]float64{
	SymbolPurpleOwl: 0.2,
	SymbolGreenOwl:  0.2,
	SymbolYellowOwl: 0.2,
	SymbolBlueOwl:   0.2,
	SymbolRedOwl:    0.2,

	SymbolFreeGame:    0.0001, // 0.5%
	SymbolOrangeSlice: 0.0001, // 0.1%
	SymbolHoneyPot:    0.0001, // 0.1%
	SymbolStrawberry:  0.0001, // 0.1%
}

// GetLevelSpecificWeights returns symbol weights for a specific level
// Stage-cleared symbols only appear on their corresponding level
func GetLevelSpecificWeights(level Level) map[Symbol]float64 {
	weights := make(map[Symbol]float64)

	// Base weights for all levels
	weights[SymbolPurpleOwl] = 0.2475
	weights[SymbolGreenOwl] = 0.2475
	weights[SymbolYellowOwl] = 0.2475
	weights[SymbolBlueOwl] = 0.2475
	weights[SymbolRedOwl] = 0.2475
	weights[SymbolFreeGame] = 0.1 // much rarer(0.002) for testing 0.1 is okay

	// Add level-specific stage-cleared symbol
	switch level {
	case Level1:
		weights[SymbolOrangeSlice] = 0.002 // much rarer(0.002) for testing 0.1 is okay
	case Level2:
		weights[SymbolHoneyPot] = 0.1 // much rarer(0.002) for testing 0.1 is okay
	case Level3:
		weights[SymbolStrawberry] = 0.1 // much rarer(0.002) for testing 0.1 is okay
	}

	return weights
}

// Paytables for each level (payouts for Bet Multiplier = 1)
// Only regular bird symbols have payouts, stage-cleared symbols don't pay

// Level 1 Paytable (4x4 grid, supports 4-16 connected symbols)
var PaytableLevel1 = map[Symbol]map[int]float64{
	SymbolPurpleOwl: {4: 2, 5: 4, 6: 5, 7: 8, 8: 10, 9: 20, 10: 30, 11: 50, 12: 100, 13: 200, 14: 400, 15: 400, 16: 400},
	SymbolGreenOwl:  {4: 4, 5: 5, 6: 10, 7: 20, 8: 30, 9: 50, 10: 100, 11: 250, 12: 500, 13: 750, 14: 800, 15: 800, 16: 800},
	SymbolYellowOwl: {4: 5, 5: 10, 6: 20, 7: 40, 8: 80, 9: 160, 10: 500, 11: 1000, 12: 2000, 13: 5000, 14: 6000, 15: 6000, 16: 6000},
	SymbolBlueOwl:   {4: 10, 5: 30, 6: 50, 7: 60, 8: 100, 9: 750, 10: 1000, 11: 10000, 12: 20000, 13: 50000, 14: 60000, 15: 60000, 16: 60000},
	SymbolRedOwl:    {4: 20, 5: 50, 6: 100, 7: 500, 8: 1000, 9: 2000, 10: 5000, 11: 20000, 12: 50000, 13: 60000, 14: 80000, 15: 80000, 16: 80000},
}

// Level 2 Paytable (5x5 grid, supports 5-25 connected symbols)
var PaytableLevel2 = map[Symbol]map[int]float64{
	SymbolPurpleOwl: {5: 2, 6: 4, 7: 5, 8: 8, 9: 10, 10: 20, 11: 30, 12: 50, 13: 100, 14: 200, 15: 450, 16: 450, 17: 450, 18: 450, 19: 450, 20: 450, 21: 450, 22: 450, 23: 450, 24: 450, 25: 450},
	SymbolGreenOwl:  {5: 4, 6: 5, 7: 10, 8: 20, 9: 30, 10: 50, 11: 100, 12: 250, 13: 500, 14: 750, 15: 1000, 16: 1000, 17: 1000, 18: 1000, 19: 1000, 20: 1000, 21: 1000, 22: 1000, 23: 1000, 24: 1000, 25: 1000},
	SymbolYellowOwl: {5: 5, 6: 10, 7: 20, 8: 40, 9: 80, 10: 160, 11: 500, 12: 1000, 13: 2000, 14: 5000, 15: 7000, 16: 7000, 17: 7000, 18: 7000, 19: 7000, 20: 7000, 21: 7000, 22: 7000, 23: 7000, 24: 7000, 25: 7000},
	SymbolBlueOwl:   {5: 10, 6: 30, 7: 50, 8: 60, 9: 100, 10: 750, 11: 1000, 12: 10000, 13: 20000, 14: 50000, 15: 70000, 16: 70000, 17: 70000, 18: 70000, 19: 70000, 20: 70000, 21: 70000, 22: 70000, 23: 70000, 24: 70000, 25: 70000},
	SymbolRedOwl:    {5: 20, 6: 50, 7: 100, 8: 500, 9: 1000, 10: 2000, 11: 5000, 12: 20000, 13: 50000, 14: 80000, 15: 100000, 16: 100000, 17: 100000, 18: 100000, 19: 100000, 20: 100000, 21: 100000, 22: 100000, 23: 100000, 24: 100000, 25: 100000},
}

// Level 3 Paytable (6x6 grid, supports 6-36 connected symbols)
var PaytableLevel3 = map[Symbol]map[int]float64{
	SymbolPurpleOwl: {6: 2, 7: 4, 8: 5, 9: 8, 10: 10, 11: 20, 12: 30, 13: 50, 14: 100, 15: 200, 16: 500, 17: 500, 18: 500, 19: 500, 20: 500, 21: 500, 22: 500, 23: 500, 24: 500, 25: 500, 26: 500, 27: 500, 28: 500, 29: 500, 30: 500, 31: 500, 32: 500, 33: 500, 34: 500, 35: 500, 36: 500},
	SymbolGreenOwl:  {6: 4, 7: 5, 8: 10, 9: 20, 10: 30, 11: 50, 12: 100, 13: 250, 14: 500, 15: 750, 16: 1200, 17: 1200, 18: 1200, 19: 1200, 20: 1200, 21: 1200, 22: 1200, 23: 1200, 24: 1200, 25: 1200, 26: 1200, 27: 1200, 28: 1200, 29: 1200, 30: 1200, 31: 1200, 32: 1200, 33: 1200, 34: 1200, 35: 1200, 36: 1200},
	SymbolYellowOwl: {6: 5, 7: 10, 8: 20, 9: 40, 10: 80, 11: 160, 12: 500, 13: 1000, 14: 2000, 15: 5000, 16: 8000, 17: 8000, 18: 8000, 19: 8000, 20: 8000, 21: 8000, 22: 8000, 23: 8000, 24: 8000, 25: 8000, 26: 8000, 27: 8000, 28: 8000, 29: 8000, 30: 8000, 31: 8000, 32: 8000, 33: 8000, 34: 8000, 35: 8000, 36: 8000},
	SymbolBlueOwl:   {6: 10, 7: 30, 8: 50, 9: 60, 10: 100, 11: 750, 12: 1000, 13: 10000, 14: 20000, 15: 50000, 16: 80000, 17: 80000, 18: 80000, 19: 80000, 20: 80000, 21: 80000, 22: 80000, 23: 80000, 24: 80000, 25: 80000, 26: 80000, 27: 80000, 28: 80000, 29: 80000, 30: 80000, 31: 80000, 32: 80000, 33: 80000, 34: 80000, 35: 80000, 36: 80000},
	SymbolRedOwl:    {6: 20, 7: 50, 8: 100, 9: 500, 10: 1000, 11: 2000, 12: 5000, 13: 20000, 14: 50000, 15: 100000, 16: 100000, 17: 100000, 18: 100000, 19: 100000, 20: 100000, 21: 100000, 22: 100000, 23: 100000, 24: 100000, 25: 100000, 26: 100000, 27: 100000, 28: 100000, 29: 100000, 30: 100000, 31: 100000, 32: 100000, 33: 100000, 34: 100000, 35: 100000, 36: 100000},
}

// GetPaytable returns the appropriate paytable for the given level
func GetPaytable(level Level) map[Symbol]map[int]float64 {
	switch level {
	case Level1:
		return PaytableLevel1
	case Level2:
		return PaytableLevel2
	case Level3:
		return PaytableLevel3
	default:
		return PaytableLevel1
	}
}

// IsStageClearedSymbol checks if a symbol is a stage-cleared symbol
func IsStageClearedSymbol(symbol Symbol) bool {
	return symbol == SymbolOrangeSlice || symbol == SymbolHoneyPot || symbol == SymbolStrawberry
}

// IsRegularBirdSymbol checks if a symbol is a regular bird symbol (can form connections)
func IsRegularBirdSymbol(symbol Symbol) bool {
	return symbol == SymbolPurpleOwl || symbol == SymbolGreenOwl ||
		symbol == SymbolYellowOwl || symbol == SymbolBlueOwl || symbol == SymbolRedOwl
}
