package birdsparty

import (
	"log"
	"math"
	"math/rand"
)

// WeightedRandomSymbol selects a symbol based on level-specific weights
func WeightedRandomSymbol(level Level, r *rand.Rand) Symbol {
	weights := GetLevelSpecificWeights(level)

	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}

	roll := r.Float64() * totalWeight
	currentWeight := 0.0
	for symbol, weight := range weights {
		currentWeight += weight
		if roll <= currentWeight {
			return symbol
		}
	}
	return SymbolPurpleOwl // Fallback
}

// GenerateGrid generates a grid of specified size with symbols for the given level
func GenerateGrid(level Level, r *rand.Rand) [][]string {
	gridSize := level.GetGridSize()
	grid := make([][]string, gridSize)
	freeGamePlaced := false
	for y := 0; y < gridSize; y++ {
		grid[y] = make([]string, gridSize)
		for x := 0; x < gridSize; x++ {
			weights := GetLevelSpecificWeights(level)
			if freeGamePlaced {
				delete(weights, SymbolFreeGame)
			}
			symbol := weightedRandomSymbolCustomWeights(weights, r)
			if symbol == SymbolFreeGame {
				freeGamePlaced = true
			}
			grid[y][x] = string(symbol)
		}
	}
	return grid
}

// Helper for custom weights (same as WeightedRandomSymbol but takes weights as argument)
func weightedRandomSymbolCustomWeights(weights map[Symbol]float64, r *rand.Rand) Symbol {
	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}
	roll := r.Float64() * totalWeight
	currentWeight := 0.0
	for symbol, weight := range weights {
		currentWeight += weight
		if roll <= currentWeight {
			return symbol
		}
	}
	return SymbolPurpleOwl // Fallback
}

// GenerateGridWithWin generates a grid that has potential connections (bird symbols only)
func GenerateGridWithWin(level Level, r *rand.Rand) [][]string {
	gridSize := level.GetGridSize()
	log.Printf("Generating grid with win for level %d with grid size %dx%d", level, gridSize, gridSize)
	maxAttempts := 100

	for attempts := 0; attempts < maxAttempts; attempts++ {
		grid := GenerateGrid(level, r)
		// Check for bird symbol connections (ignore stage-cleared symbols)
		connections := FindRegularConnections(grid, level)
		if len(connections) > 0 {
			return grid
		}
	}

	// If we can't generate a natural win, force one
	return ForceWinGrid(level, r)
}

// GenerateLossGrid generates a grid with no winning connections (bird symbols)
func GenerateLossGrid(level Level, r *rand.Rand) [][]string {
	gridSize := level.GetGridSize()
	log.Printf("Generating loss grid for level %d with grid size %dx%d", level, gridSize, gridSize)
	maxAttempts := 100

	for attempts := 0; attempts < maxAttempts; attempts++ {
		grid := GenerateGrid(level, r)
		// Check for bird symbol connections (ignore stage-cleared symbols)
		connections := FindRegularConnections(grid, level)
		if len(connections) == 0 {
			return grid
		}
	}

	// If we can't generate a natural loss, force one
	return ForceLossGrid(level, r)
}

// ForceWinGrid creates a grid with guaranteed bird symbol connections
func ForceWinGrid(level Level, r *rand.Rand) [][]string {
	gridSize := level.GetGridSize()
	grid := GenerateGrid(level, r)
	minConnection := level.GetMinConnection()

	// Pick a random bird symbol
	birdSymbols := []Symbol{SymbolPurpleOwl, SymbolGreenOwl, SymbolYellowOwl, SymbolBlueOwl, SymbolRedOwl}
	targetSymbol := birdSymbols[r.Intn(len(birdSymbols))]

	// Create a horizontal line of the minimum required length
	startX := r.Intn(gridSize - minConnection + 1)
	y := r.Intn(gridSize)

	for i := 0; i < minConnection; i++ {
		grid[y][startX+i] = string(targetSymbol)
	}

	return grid
}

// ForceLossGrid creates a grid with no bird symbol connections
func ForceLossGrid(level Level, r *rand.Rand) [][]string {
	gridSize := level.GetGridSize()
	grid := make([][]string, gridSize)
	birdSymbols := []Symbol{SymbolPurpleOwl, SymbolGreenOwl, SymbolYellowOwl, SymbolBlueOwl, SymbolRedOwl}

	for y := 0; y < gridSize; y++ {
		grid[y] = make([]string, gridSize)
		for x := 0; x < gridSize; x++ {
			// Ensure no consecutive bird symbols
			availableSymbols := make([]Symbol, len(birdSymbols))
			copy(availableSymbols, birdSymbols)

			// Remove symbols that would create connections
			if x > 0 && grid[y][x-1] != "" && IsRegularBirdSymbol(Symbol(grid[y][x-1])) {
				prevSymbol := Symbol(grid[y][x-1])
				for i, sym := range availableSymbols {
					if sym == prevSymbol {
						availableSymbols = append(availableSymbols[:i], availableSymbols[i+1:]...)
						break
					}
				}
			}

			if y > 0 && grid[y-1][x] != "" && IsRegularBirdSymbol(Symbol(grid[y-1][x])) {
				upSymbol := Symbol(grid[y-1][x])
				for i, sym := range availableSymbols {
					if sym == upSymbol {
						availableSymbols = append(availableSymbols[:i], availableSymbols[i+1:]...)
						break
					}
				}
			}

			if len(availableSymbols) > 0 {
				grid[y][x] = string(availableSymbols[r.Intn(len(availableSymbols))])
			} else {
				grid[y][x] = string(birdSymbols[r.Intn(len(birdSymbols))])
			}
		}
	}

	return grid
}

// FindStageClearedSymbols finds all stage-cleared symbols for the current level
func FindStageClearedSymbols(grid [][]string, level Level) []StageClearedSymbol {
	var stageClearedSymbols []StageClearedSymbol
	gridSize := len(grid)
	expectedSymbol := level.GetStageClearedSymbol()

	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			if grid[y][x] == string(expectedSymbol) {
				stageClearedSymbols = append(stageClearedSymbols, StageClearedSymbol{
					Symbol:   expectedSymbol,
					Position: Position{X: x, Y: y},
				})
			}
		}
	}

	log.Printf("Found %d stage-cleared symbols (%s) for level %d",
		len(stageClearedSymbols), expectedSymbol, level)

	return stageClearedSymbols
}

// RemoveStageClearedSymbols removes all stage-cleared symbols from the grid
func RemoveStageClearedSymbols(grid [][]string, stageClearedSymbols []StageClearedSymbol) {
	for _, stageSymbol := range stageClearedSymbols {
		pos := stageSymbol.Position
		if pos.X >= 0 && pos.X < len(grid) && pos.Y >= 0 && pos.Y < len(grid) {
			grid[pos.Y][pos.X] = ""
			log.Printf("Removed stage-cleared symbol %s at position (%d,%d)",
				stageSymbol.Symbol, pos.X, pos.Y)
		}
	}
}

// FindRegularConnections finds all bird symbol connections in the grid (excludes stage-cleared symbols)
func FindRegularConnections(grid [][]string, level Level) []Connection {
	var connections []Connection
	gridSize := len(grid)
	visited := make([][]bool, gridSize)
	for i := range visited {
		visited[i] = make([]bool, gridSize)
	}

	minConnection := level.GetMinConnection()

	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			if !visited[y][x] && IsRegularBirdSymbol(Symbol(grid[y][x])) {
				symbol := Symbol(grid[y][x])
				positions := findConnectedPositions(grid, x, y, symbol, visited)

				if len(positions) >= minConnection {
					payout := calculatePayout(symbol, len(positions), level, 1) // Base multiplier
					connections = append(connections, Connection{
						Symbol:    symbol,
						Positions: positions,
						Count:     len(positions),
						Payout:    payout,
					})
				}
			}
		}
	}

	return connections
}

// findConnectedPositions uses flood fill to find all connected positions (bird symbols only)
func findConnectedPositions(grid [][]string, startX, startY int, symbol Symbol, visited [][]bool) []Position {
	var positions []Position
	var stack []Position
	gridSize := len(grid)

	stack = append(stack, Position{X: startX, Y: startY})

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if current.X < 0 || current.X >= gridSize || current.Y < 0 || current.Y >= gridSize {
			continue
		}

		if visited[current.Y][current.X] ||
			grid[current.Y][current.X] != string(symbol) ||
			!IsRegularBirdSymbol(Symbol(grid[current.Y][current.X])) {
			continue
		}

		visited[current.Y][current.X] = true
		positions = append(positions, current)

		// Add adjacent positions (horizontal and vertical only)
		stack = append(stack, Position{X: current.X + 1, Y: current.Y})
		stack = append(stack, Position{X: current.X - 1, Y: current.Y})
		stack = append(stack, Position{X: current.X, Y: current.Y + 1})
		stack = append(stack, Position{X: current.X, Y: current.Y - 1})
	}

	return positions
}

// calculatePayout calculates the payout for a connection
func calculatePayout(symbol Symbol, count int, level Level, betMultiplier int) float64 {
	paytable := GetPaytable(level)

	if payoutMap, exists := paytable[symbol]; exists {
		if payout, found := payoutMap[count]; found {
			result := payout * 0.01 * float64(betMultiplier) // Denomination is 0.01
			return math.Round(result*100) / 100
		}
	}

	return 0
}

// RemoveConnections removes connected symbols from the grid and returns positions to fill
func RemoveConnections(grid [][]string, connections []Connection) []Position {
	var removedPositions []Position

	for _, connection := range connections {
		for _, pos := range connection.Positions {
			if pos.X >= 0 && pos.X < len(grid) && pos.Y >= 0 && pos.Y < len(grid) {
				grid[pos.Y][pos.X] = ""
				removedPositions = append(removedPositions, pos)
			}
		}
	}

	return removedPositions
}

// ApplyGravity makes symbols fall down to fill empty spaces
func ApplyGravity(grid [][]string, level Level, r *rand.Rand) {
	gridSize := len(grid)

	for x := 0; x < gridSize; x++ {
		// Move existing symbols down
		writePos := gridSize - 1
		for y := gridSize - 1; y >= 0; y-- {
			if grid[y][x] != "" {
				if y != writePos {
					grid[writePos][x] = grid[y][x]
					grid[y][x] = ""
				}
				writePos--
			}
		}

		// Fill empty spaces at the top with new symbols
		for y := 0; y <= writePos; y++ {
			grid[y][x] = string(WeightedRandomSymbol(level, r))
		}
	}
}

// CountFreeGameSymbols counts free game symbols in the grid
func CountFreeGameSymbols(grid [][]string) int {
	count := 0
	gridSize := len(grid)
	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			if grid[y][x] == string(SymbolFreeGame) {
				count++
			}
		}
	}
	return count
}

// GetRandomFreeSpinMultiplier returns a random multiplier between 1.0 and 5.0
func GetRandomFreeSpinMultiplier(r *rand.Rand) float64 {
	multipliers := []float64{1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0}
	return multipliers[r.Intn(len(multipliers))]
}

// AdvanceLevel advances to the next level or returns to level 1 after level 3
func AdvanceLevel(currentLevel Level) Level {
	switch currentLevel {
	case Level1:
		return Level2
	case Level2:
		return Level3
	case Level3:
		return Level1
	default:
		return Level1
	}
}

// InitializeGameState initializes a new game state with default values
func InitializeGameState() GameState {
	return GameState{
		CurrentLevel:  Level1,
		GridSize:      Level1GridSize,
		Grid:          [][]string{},
		StageProgress: 0,
		GameMode:      "base",
		FreeSpins: struct {
			Remaining    int     `json:"remaining"`
			TotalAwarded int     `json:"totalAwarded"`
			Multiplier   float64 `json:"multiplier"`
		}{
			Remaining:    0,
			TotalAwarded: 0,
			Multiplier:   1.0,
		},
		TotalWin:            0,
		Cascading:           false,
		LastConnections:     []Connection{},
		CascadeCount:        0,
		StageClearedSymbols: []StageClearedSymbol{},
	}
}

// UpdateGameStateForLevel updates the game state when advancing to a new level
func UpdateGameStateForLevel(gameState *GameState, newLevel Level) {
	gameState.CurrentLevel = newLevel
	gameState.GridSize = newLevel.GetGridSize()
	gameState.StageProgress = 0 // Reset progress for new level

	log.Printf("Advanced to Level %d with %dx%d grid", newLevel, gameState.GridSize, gameState.GridSize)
}

// ProcessStageClearedSymbols processes stage-cleared symbols and checks for level advancement
func ProcessStageClearedSymbols(gameState *GameState, stageClearedSymbols []StageClearedSymbol, level Level, r *rand.Rand) (bool, Level, Level) {
	if len(stageClearedSymbols) == 0 {
		return false, gameState.CurrentLevel, gameState.CurrentLevel
	}

	oldLevel := gameState.CurrentLevel

	// Remove stage-cleared symbols from grid
	RemoveStageClearedSymbols(gameState.Grid, stageClearedSymbols)

	// Apply gravity after removing stage-cleared symbols
	ApplyGravity(gameState.Grid, level, r)

	// Update stage progress
	gameState.StageProgress += len(stageClearedSymbols)
	log.Printf("Added %d stage-cleared symbols to progress, total: %d/15",
		len(stageClearedSymbols), gameState.StageProgress)

	// Check for level advancement
	levelAdvanced := false
	if gameState.StageProgress >= StageProgressTarget {
		newLevel := AdvanceLevel(oldLevel)

		// Handle overflow progress
		excessProgress := gameState.StageProgress - StageProgressTarget

		UpdateGameStateForLevel(gameState, newLevel)

		// Carry over excess progress to new level
		gameState.StageProgress = excessProgress

		// Regenerate grid with new level's size and symbols
		gameState.Grid = GenerateGrid(newLevel, r)

		levelAdvanced = true
		log.Printf("Level advanced from %d to %d, excess progress: %d", oldLevel, newLevel, excessProgress)

		return levelAdvanced, oldLevel, newLevel
	}

	return levelAdvanced, oldLevel, oldLevel
}

// ValidateGridDimensions ensures grid matches expected size for level
func ValidateGridDimensions(grid [][]string, level Level) bool {
	expectedSize := level.GetGridSize()
	if len(grid) != expectedSize {
		return false
	}
	for _, row := range grid {
		if len(row) != expectedSize {
			return false
		}
	}
	return true
}

// CleanupInvalidSymbols removes any invalid symbols that don't belong to current level
func CleanupInvalidSymbols(grid [][]string, level Level, r *rand.Rand) {
	gridSize := len(grid)
	levelStageClearedSymbol := level.GetStageClearedSymbol()

	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			symbol := Symbol(grid[y][x])

			// If it's a stage-cleared symbol that doesn't belong to current level, replace it
			if IsStageClearedSymbol(symbol) && symbol != levelStageClearedSymbol {
				newSymbol := WeightedRandomSymbol(level, r)
				grid[y][x] = string(newSymbol)
				log.Printf("Replaced invalid stage-cleared symbol %s with %s at (%d,%d)",
					symbol, newSymbol, x, y)
			}
		}
	}
}

// HasPotentialConnections checks if grid has any potential bird symbol connections
func HasPotentialConnections(grid [][]string, level Level) bool {
	connections := FindRegularConnections(grid, level)
	return len(connections) > 0
}

// CountStageClearedSymbolsInGrid counts how many stage-cleared symbols are currently in the grid
func CountStageClearedSymbolsInGrid(grid [][]string, level Level) int {
	stageClearedSymbols := FindStageClearedSymbols(grid, level)
	return len(stageClearedSymbols)
}
