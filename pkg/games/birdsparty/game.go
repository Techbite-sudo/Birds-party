package birdsparty

import (
	"fmt"
	"log"
	"math"
	"math/rand"
)

// round rounds a float64 to two decimal places
func round(val float64) float64 {
	return math.Round(val*100) / 100
}

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

// Helper: Checks if SymbolFreeGame is already present in the grid
func hasFreeGameSymbol(grid [][]string) bool {
	for y := range grid {
		for x := range grid[y] {
			if grid[y][x] == string(SymbolFreeGame) {
				return true
			}
		}
	}
	return false
}

// Modified WeightedRandomSymbol to accept allowFreeGame argument
// WeightedRandomSymbolWithControl now takes a forbidFreeGame argument (true = never allow free_game)
func WeightedRandomSymbolWithControl(level Level, r *rand.Rand, forbidFreeGame bool) Symbol {
	weights := GetLevelSpecificWeights(level)
	if forbidFreeGame {
		delete(weights, SymbolFreeGame)
	}

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
// If forbidFreeGame is true, free_game symbol will never appear
func GenerateGrid(level Level, r *rand.Rand, forbidFreeGame bool) [][]string {
	gridSize := level.GetGridSize()
	grid := make([][]string, gridSize)
	freeGamePlaced := false
	for y := 0; y < gridSize; y++ {
		grid[y] = make([]string, gridSize)
		for x := 0; x < gridSize; x++ {
			allowFreeGame := !freeGamePlaced && !forbidFreeGame
			symbol := WeightedRandomSymbolWithControl(level, r, !allowFreeGame)
			if symbol == SymbolFreeGame {
				freeGamePlaced = true
			}
			grid[y][x] = string(symbol)
		}
	}
	return grid
}

// GenerateGridWithWin generates a grid that has potential connections (bird symbols only)
// If forbidFreeGame is true, free_game symbol will never appear
func GenerateGridWithWin(level Level, r *rand.Rand, forbidFreeGame bool) [][]string {
	gridSize := level.GetGridSize()
	log.Printf("Generating grid with win for level %d with grid size %dx%d", level, gridSize, gridSize)
	maxAttempts := 100

	for attempts := 0; attempts < maxAttempts; attempts++ {
		grid := GenerateGrid(level, r, forbidFreeGame)
		// Check for bird symbol connections (ignore stage-cleared symbols)
		connections := FindRegularConnections(grid, level)
		if len(connections) > 0 {
			return grid
		}
	}

	// If we can't generate a natural win, force one
	return ForceWinGrid(level, r, forbidFreeGame)
}

// GenerateLossGrid generates a grid with no winning connections (bird symbols)
// If forbidFreeGame is true, free_game symbol will never appear
func GenerateLossGrid(level Level, r *rand.Rand, forbidFreeGame bool) [][]string {
	gridSize := level.GetGridSize()
	log.Printf("Generating loss grid for level %d with grid size %dx%d", level, gridSize, gridSize)
	maxAttempts := 100

	for attempts := 0; attempts < maxAttempts; attempts++ {
		grid := GenerateGrid(level, r, forbidFreeGame)
		// Check for bird symbol connections (ignore stage-cleared symbols)
		connections := FindRegularConnections(grid, level)
		if len(connections) == 0 {
			return grid
		}
	}

	// If we can't generate a natural loss, force one
	return ForceLossGrid(level, r, forbidFreeGame)
}

// ForceWinGrid creates a grid with guaranteed bird symbol connections
// If forbidFreeGame is true, free_game symbol will never appear
func ForceWinGrid(level Level, r *rand.Rand, forbidFreeGame bool) [][]string {
	gridSize := level.GetGridSize()
	grid := GenerateGrid(level, r, forbidFreeGame)
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
// If forbidFreeGame is true, free_game symbol will never appear
func ForceLossGrid(level Level, r *rand.Rand, forbidFreeGame bool) [][]string {
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

	// Remove free_game symbol if forbidden
	if forbidFreeGame {
		for y := 0; y < gridSize; y++ {
			for x := 0; x < gridSize; x++ {
				if grid[y][x] == string(SymbolFreeGame) {
					grid[y][x] = string(birdSymbols[r.Intn(len(birdSymbols))])
				}
			}
		}
	}

	return grid
}

// ProcessStageClearedSymbolsSurgical processes stage-cleared symbols with surgical precision
// This preserves the grid structure and only affects the stage-cleared symbol positions
func ProcessStageClearedSymbolsSurgical(gameState *GameState, stageClearedSymbols []StageClearedSymbol, level Level, r *rand.Rand) (bool, Level, Level) {
	if len(stageClearedSymbols) == 0 {
		return false, gameState.CurrentLevel, gameState.CurrentLevel
	}

	oldLevel := gameState.CurrentLevel

	// Remove stage-cleared symbols from grid SURGICALLY
	RemoveStageClearedSymbolsSurgical(gameState.Grid, stageClearedSymbols)

	// Apply gravity SURGICALLY - only affects columns with removed symbols
	ApplyGravitySurgical(gameState.Grid, stageClearedSymbols, level, r, false)

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
		gameState.Grid = GenerateGrid(newLevel, r, false) // No free game in new level

		levelAdvanced = true
		log.Printf("Level advanced from %d to %d, excess progress: %d", oldLevel, newLevel, excessProgress)

		return levelAdvanced, oldLevel, newLevel
	}

	return levelAdvanced, oldLevel, oldLevel
}

// RemoveStageClearedSymbolsSurgical removes only stage-cleared symbols, preserving grid structure
func RemoveStageClearedSymbolsSurgical(grid [][]string, stageClearedSymbols []StageClearedSymbol) {
	for _, stageSymbol := range stageClearedSymbols {
		pos := stageSymbol.Position
		if pos.X >= 0 && pos.X < len(grid) && pos.Y >= 0 && pos.Y < len(grid) {
			grid[pos.Y][pos.X] = ""
			log.Printf("Surgically removed stage-cleared symbol %s at position (%d,%d)",
				stageSymbol.Symbol, pos.X, pos.Y)
		}
	}
}

// ApplyGravitySurgical applies gravity only to columns affected by stage-cleared symbol removal
// If forbidFreeGame is true, free_game symbol will never appear
func ApplyGravitySurgical(grid [][]string, stageClearedSymbols []StageClearedSymbol, level Level, r *rand.Rand, forbidFreeGame bool) []Position {
	gridSize := len(grid)
	var newPositions []Position

	// Get unique columns that need gravity applied
	affectedColumns := make(map[int]bool)
	for _, stageSymbol := range stageClearedSymbols {
		affectedColumns[stageSymbol.Position.X] = true
	}

	log.Printf("Applying surgical gravity to columns: %v", getKeys(affectedColumns))

	// Apply gravity only to affected columns
	for x := range affectedColumns {
		if x >= 0 && x < gridSize {
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
				allowFreeGame := !hasFreeGameSymbol(grid) && !forbidFreeGame
				grid[y][x] = string(WeightedRandomSymbolWithControl(level, r, !allowFreeGame))
				log.Printf("Generated new symbol %s at position (%d,%d) after surgical gravity", grid[y][x], x, y)
				newPositions = append(newPositions, Position{X: x, Y: y})
			}
		}
	}
	return newPositions
}

// ApplySurgicalLoss attempts to remove connections while preserving the grid structure
// Only modifies newly generated positions
// Returns true if surgical loss was successful, false if impossible
func ApplySurgicalLoss(gameState *GameState, originalGrid [][]string, stageClearedSymbols []StageClearedSymbol, level Level, r *rand.Rand, newPositions []Position) bool {
	// Build a set of allowed positions for modification
	allowed := make(map[string]bool)
	for _, pos := range newPositions {
		key := fmt.Sprintf("%d,%d", pos.X, pos.Y)
		allowed[key] = true
	}

	connections := FindRegularConnections(gameState.Grid, level)
	if len(connections) == 0 {
		// No connections to remove, surgical loss already achieved
		return true
	}

	log.Printf("Attempting surgical loss on %d connections (only new positions)", len(connections))

	maxAttempts := 50
	for attempts := 0; attempts < maxAttempts; attempts++ {
		// Create a copy of current grid
		testGrid := make([][]string, len(gameState.Grid))
		for i := range gameState.Grid {
			testGrid[i] = make([]string, len(gameState.Grid[i]))
			copy(testGrid[i], gameState.Grid[i])
		}

		// Try modifying a few allowed positions to break connections
		modificationsCount := min(3, len(allowed))
		modified := 0

		for posKey := range allowed {
			if modified >= modificationsCount {
				break
			}

			// Parse position
			var x, y int
			fmt.Sscanf(posKey, "%d,%d", &x, &y)

			if x >= 0 && x < len(testGrid) && y >= 0 && y < len(testGrid[0]) {
				originalSymbol := testGrid[y][x]
				// Respect free spins mode - don't allow free game symbols during free spins
				forbidFreeGame := gameState.GameMode == "freeSpins"
				newSymbol := WeightedRandomSymbolWithControl(level, r, forbidFreeGame)
				testGrid[y][x] = string(newSymbol)

				// Check if this breaks connections
				testConnections := FindRegularConnections(testGrid, level)
				if len(testConnections) == 0 {
					// Success! Apply this modification
					gameState.Grid = testGrid
					log.Printf("Surgical loss successful: changed symbol at (%d,%d) from %s to %s (new only)", x, y, originalSymbol, newSymbol)
					return true
				}

				modified++
			}
		}
	}

	log.Printf("Surgical loss impossible: stage-cleared processing created unbreakable winning configuration (new only)")
	return false
}

// Helper function to get keys from map
func getKeys(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// RemoveConnectionsSurgical removes connected symbols from the grid and returns affected positions
// This tracks which positions were removed for surgical gravity application
func RemoveConnectionsSurgical(grid [][]string, connections []Connection) []Position {
	var affectedPositions []Position

	for _, connection := range connections {
		for _, pos := range connection.Positions {
			if pos.X >= 0 && pos.X < len(grid) && pos.Y >= 0 && pos.Y < len(grid) {
				grid[pos.Y][pos.X] = ""
				affectedPositions = append(affectedPositions, pos)
				log.Printf("Surgically removed %s symbol at position (%d,%d)",
					connection.Symbol, pos.X, pos.Y)
			}
		}
	}

	return affectedPositions
}

// ApplyGravitySurgicalForCascade applies gravity only to columns affected by connection removal
// If forbidFreeGame is true, free_game symbol will never appear
func ApplyGravitySurgicalForCascade(grid [][]string, affectedPositions []Position, level Level, r *rand.Rand, forbidFreeGame bool) []Position {
	gridSize := len(grid)
	var newPositions []Position

	// Get unique columns that need gravity applied
	affectedColumns := make(map[int]bool)
	for _, pos := range affectedPositions {
		affectedColumns[pos.X] = true
	}

	log.Printf("Applying surgical cascade gravity to columns: %v", getKeys(affectedColumns))

	// Apply gravity only to affected columns
	for x := range affectedColumns {
		if x >= 0 && x < gridSize {
			// Move existing symbols down
			writePos := gridSize - 1
			for y := gridSize - 1; y >= 0; y-- {
				if grid[y][x] != "" {
					if y != writePos {
						grid[writePos][x] = grid[y][x]
						grid[y][x] = ""
						log.Printf("Moved symbol %s from (%d,%d) to (%d,%d) via cascade gravity", grid[writePos][x], x, y, x, writePos)
					}
					writePos--
				}
			}

			// Fill empty spaces at the top with new symbols
			for y := 0; y <= writePos; y++ {
				allowFreeGame := !hasFreeGameSymbol(grid) && !forbidFreeGame
				grid[y][x] = string(WeightedRandomSymbolWithControl(level, r, !allowFreeGame))
				log.Printf("Generated new symbol %s at position (%d,%d) after cascade gravity", grid[y][x], x, y)
				newPositions = append(newPositions, Position{X: x, Y: y})
			}
		}
	}
	return newPositions
}

// ApplySurgicalLossForCascade attempts to remove connections while preserving the grid structure for cascades
// Only modifies newly generated positions
// Returns true if surgical loss was successful, false if impossible
func ApplySurgicalLossForCascade(gameState *GameState, originalGrid [][]string, newPositions []Position, level Level, r *rand.Rand) bool {
	// Build a set of allowed positions for modification
	allowed := make(map[string]bool)
	for _, pos := range newPositions {
		key := fmt.Sprintf("%d,%d", pos.X, pos.Y)
		allowed[key] = true
	}

	connections := FindRegularConnections(gameState.Grid, level)
	if len(connections) == 0 {
		// No connections to remove, surgical loss already achieved
		return true
	}

	log.Printf("Attempting surgical cascade loss on %d connections (only new positions)", len(connections))

	maxAttempts := 50
	for attempts := 0; attempts < maxAttempts; attempts++ {
		// Create a copy of current grid
		testGrid := make([][]string, len(gameState.Grid))
		for i := range gameState.Grid {
			testGrid[i] = make([]string, len(gameState.Grid[i]))
			copy(testGrid[i], gameState.Grid[i])
		}

		// Try modifying a few allowed positions to break connections
		modificationsCount := min(4, len(allowed))
		modified := 0

		for posKey := range allowed {
			if modified >= modificationsCount {
				break
			}

			// Parse position
			var x, y int
			fmt.Sscanf(posKey, "%d,%d", &x, &y)

			if x >= 0 && x < len(testGrid) && y >= 0 && y < len(testGrid[0]) {
				originalSymbol := testGrid[y][x]
				// Respect free spins mode - don't allow free game symbols during free spins
				forbidFreeGame := gameState.GameMode == "freeSpins"
				newSymbol := WeightedRandomSymbolWithControl(level, r, forbidFreeGame)
				testGrid[y][x] = string(newSymbol)

				// Check if this breaks connections
				testConnections := FindRegularConnections(testGrid, level)
				if len(testConnections) == 0 {
					// Success! Apply this modification
					gameState.Grid = testGrid
					log.Printf("Surgical cascade loss successful: changed symbol at (%d,%d) from %s to %s (new only)", x, y, originalSymbol, newSymbol)
					return true
				}

				modified++
			}
		}
	}

	log.Printf("Surgical cascade loss impossible: cascade processing created unbreakable winning configuration (new only)")
	return false
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// RemoveStageClearedSymbols removes all stage-cleared symbols from the grid (LEGACY - use surgical version)
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
			return round(result)
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

// ApplyGravity makes symbols fall down to fill empty spaces (LEGACY - use surgical version when appropriate)
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
			allowFreeGame := !hasFreeGameSymbol(grid)
			grid[y][x] = string(WeightedRandomSymbolWithControl(level, r, !allowFreeGame))
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

// ProcessStageClearedSymbols processes stage-cleared symbols and checks for level advancement (LEGACY VERSION)
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
		// Respect free spins mode - don't allow free game symbols during free spins
		forbidFreeGame := gameState.GameMode == "freeSpins"
		gameState.Grid = GenerateGrid(newLevel, r, forbidFreeGame)

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
				// Note: This function doesn't have access to gameState, so we can't check game mode
				// For now, we'll use the basic symbol generation - this function is rarely used
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
