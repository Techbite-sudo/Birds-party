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
	weights[SymbolFreeGame] = 0.001 // much rarer(0.001) for testing 0.1 is okay

	// Add level-specific stage-cleared symbol
	switch level {
	case Level1:
		weights[SymbolOrangeSlice] = 0.002 // much rarer(0.002) for testing 0.1 is okay
	case Level2:
		weights[SymbolHoneyPot] = 0.002 // much rarer(0.002) for testing 0.1 is okay
	case Level3:
		weights[SymbolStrawberry] = 0.002 // much rarer(0.002) for testing 0.1 is okay
	}

	return weights
}

// Paytables for each level (payouts for Bet Multiplier = 1)
// Only regular bird symbols have payouts, stage-cleared symbols don't pay

// Level 1 Paytable (4x4 grid, supports 4-16 connected symbols)
var PaytableLevel1 = map[Symbol]map[int]float64{
	SymbolPurpleOwl: {4: 4, 5: 5, 6: 6, 7: 7, 8: 8, 9: 9, 10: 10, 11: 11, 12: 12, 13: 13, 14: 14, 15: 15, 16: 16},
	SymbolGreenOwl:  {4: 2, 5: 4, 6: 5, 7: 8, 8: 10, 9: 20, 10: 30, 11: 50, 12: 100, 13: 200, 14: 400, 15: 800, 16: 1600},
	SymbolYellowOwl: {4: 4, 5: 5, 6: 10, 7: 20, 8: 30, 9: 50, 10: 100, 11: 250, 12: 500, 13: 750, 14: 800, 15: 1200, 16: 6000},
	SymbolBlueOwl:   {4: 5, 5: 10, 6: 20, 7: 40, 8: 60, 9: 80, 10: 160, 11: 500, 12: 1000, 13: 2000, 14: 5000, 15: 7000, 16: 8000},
	SymbolRedOwl:    {4: 10, 5: 30, 6: 50, 7: 60, 8: 100, 9: 750, 10: 1000, 11: 10000, 12: 20000, 13: 50000, 14: 60000, 15: 80000, 16: 100000},
}

// Level 2 Paytable (5x5 grid, supports 5-25 connected symbols)
var PaytableLevel2 = map[Symbol]map[int]float64{
	SymbolPurpleOwl: {5: 2, 6: 4, 7: 5, 8: 8, 9: 10, 10: 20, 11: 30, 12: 50, 13: 100, 14: 200, 15: 450, 16: 1000, 17: 1200, 18: 1500, 19: 2000, 20: 2500, 21: 3000, 22: 4000, 23: 5000, 24: 7500, 25: 10000},
	SymbolGreenOwl:  {5: 4, 6: 5, 7: 10, 8: 20, 9: 30, 10: 50, 11: 100, 12: 250, 13: 500, 14: 750, 15: 1000, 16: 7000, 17: 8000, 18: 10000, 19: 12000, 20: 15000, 21: 18000, 22: 22000, 23: 27000, 24: 32000, 25: 40000},
	SymbolYellowOwl: {5: 5, 6: 10, 7: 20, 8: 40, 9: 60, 10: 80, 11: 160, 12: 500, 13: 1000, 14: 2000, 15: 5000, 16: 8000, 17: 10000, 18: 12000, 19: 15000, 20: 20000, 21: 25000, 22: 30000, 23: 40000, 24: 50000, 25: 60000},
	SymbolBlueOwl:   {5: 10, 6: 30, 7: 50, 8: 60, 9: 100, 10: 750, 11: 1000, 12: 10000, 13: 20000, 14: 50000, 15: 70000, 16: 100000, 17: 120000, 18: 150000, 19: 180000, 20: 220000, 21: 270000, 22: 320000, 23: 400000, 24: 500000, 25: 600000},
	SymbolRedOwl:    {5: 20, 6: 50, 7: 100, 8: 500, 9: 1000, 10: 2000, 11: 5000, 12: 20000, 13: 50000, 14: 80000, 15: 100000, 16: 150000, 17: 200000, 18: 300000, 19: 400000, 20: 500000, 21: 600000, 22: 800000, 23: 1000000, 24: 1200000, 25: 1500000},
}

// Level 3 Paytable (6x6 grid, supports 6-36 connected symbols)
var PaytableLevel3 = map[Symbol]map[int]float64{
	SymbolPurpleOwl: {6: 2, 7: 4, 8: 5, 9: 8, 10: 10, 11: 20, 12: 30, 13: 50, 14: 100, 15: 200, 16: 500, 17: 600, 18: 750, 19: 900, 20: 1100, 21: 1300, 22: 1600, 23: 2000, 24: 2500, 25: 3000, 26: 3600, 27: 4300, 28: 5100, 29: 6000, 30: 7500, 31: 9000, 32: 11000, 33: 13000, 34: 16000, 35: 20000, 36: 25000},
	SymbolGreenOwl:  {6: 4, 7: 5, 8: 10, 9: 20, 10: 30, 11: 50, 12: 100, 13: 250, 14: 500, 15: 1000, 16: 8000, 17: 9000, 18: 10500, 19: 12000, 20: 14000, 21: 16000, 22: 19000, 23: 22000, 24: 26000, 25: 30000, 26: 35000, 27: 40000, 28: 46000, 29: 53000, 30: 60000, 31: 68000, 32: 77000, 33: 87000, 34: 98000, 35: 110000, 36: 125000},
	SymbolYellowOwl: {6: 5, 7: 10, 8: 20, 9: 40, 10: 60, 11: 80, 12: 160, 13: 500, 14: 1000, 15: 5000, 16: 10000, 17: 12000, 18: 14000, 19: 17000, 20: 20000, 21: 24000, 22: 28000, 23: 33000, 24: 39000, 25: 45000, 26: 52000, 27: 60000, 28: 69000, 29: 79000, 30: 90000, 31: 102000, 32: 115000, 33: 130000, 34: 146000, 35: 165000, 36: 185000},
	SymbolBlueOwl:   {6: 10, 7: 30, 8: 50, 9: 60, 10: 100, 11: 750, 12: 1000, 13: 10000, 14: 20000, 15: 50000, 16: 100000, 17: 115000, 18: 130000, 19: 150000, 20: 170000, 21: 195000, 22: 220000, 23: 250000, 24: 280000, 25: 315000, 26: 355000, 27: 400000, 28: 450000, 29: 505000, 30: 565000, 31: 630000, 32: 700000, 33: 775000, 34: 860000, 35: 950000, 36: 1050000},
	SymbolRedOwl:    {6: 20, 7: 50, 8: 100, 9: 500, 10: 1000, 11: 2000, 12: 5000, 13: 20000, 14: 50000, 15: 100000, 16: 200000, 17: 230000, 18: 265000, 19: 305000, 20: 350000, 21: 400000, 22: 460000, 23: 530000, 24: 610000, 25: 700000, 26: 800000, 27: 920000, 28: 1060000, 29: 1220000, 30: 1400000, 31: 1600000, 32: 1840000, 33: 2120000, 34: 2440000, 35: 2800000, 36: 3200000},
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
