# Birds Party Game - API Integration Guide for Unity Developers

## Overview

This document provides comprehensive guidelines for integrating the Birds Party game backend API with a Unity frontend. The game features a **dynamic grid progression system** (4x4 → 5x5 → 6x6), **stage-cleared symbol priority removal**, cluster-based connections, level advancement mechanics, cascading systems, and a Free Spin Bonus feature.

## API Endpoints

- Base URL: `https://b.api.ibibe.africa`
- Spin endpoint: `POST /spin/birdsparty`
- Stage-cleared processing: `POST /process-stage-cleared/birdsparty`
- Cascade endpoint: `POST /cascade/birdsparty`
- Health check: `GET /status`

## Game Mechanics

### Core Game Rules
- **Dynamic grid structure**: 4x4 → 5x5 → 6x6 based on level progression
- **Stage-cleared symbol priority removal**: Special symbols removed first before connections
- Cluster-based connections (horizontal/vertical adjacent symbols only)
- 3-level progression system with automatic grid expansion
- Cascading mechanics with symbol removal and gravity
- Denomination: 0.01
- Bet amounts: 0.1, 0.2, 0.3, 0.5, 1.0 (corresponding to multipliers: 1, 2, 3, 5, 10)
- Minimum bet: 10 credits per bet multiplier

### Symbols

#### Regular Bird Symbols (Form Connections)
- Purple Owl, Green Owl, Yellow Owl, Blue Owl, Red Owl (18.5% each)
- Free Game symbol (5% probability - triggers free spins)

#### Stage-Cleared Symbols (Priority Removal)
- **Level 1**: `orange_slice` - Orange slice symbol (2.5% probability on 4x4 grid)
- **Level 2**: `honey_pot` - Honey pot symbol (2.5% probability on 5x5 grid)  
- **Level 3**: `strawberry` - Strawberry symbol (2.5% probability on 6x6 grid)

**IMPORTANT**: Stage-cleared symbols do NOT form connections. They are removed individually when they appear.

### Dynamic Grid & Level System
- **Level 1**: 4x4 grid (16 positions), minimum 4 connected bird symbols required
- **Level 2**: 5x5 grid (25 positions), minimum 5 connected bird symbols required  
- **Level 3**: 6x6 grid (36 positions), minimum 6 connected bird symbols required
- **Progression**: Accumulate 15 stage-cleared symbols to advance to next level
- **Cycling**: After Level 3, returns to Level 1 (infinite progression)
- **Grid Expansion**: Grid automatically resizes when advancing levels

### Stage-Cleared Symbol Mechanics

#### Priority Removal System
1. **Stage-cleared symbols are detected** when grid is generated
2. **Priority removal**: Stage-cleared symbols are removed FIRST (ignore all connections)
3. **Gravity applied**: Symbols fall down, new symbols generated at top
4. **Then check connections**: Regular bird symbol connections checked on new grid
5. **Count toward progress**: Each removed stage-cleared symbol counts toward 15

#### Level-Specific Appearance
- **Only level-appropriate symbols appear**: Orange slice only on Level 1, etc.
- **Individual removal**: Each stage-cleared symbol is removed separately
- **No connection rules**: Stage-cleared symbols don't need to be connected
- **Progress tracking**: Each removed symbol = +1 toward level advancement

### Three-Endpoint Game Flow

#### 1. Spin Phase - `/spin/birdsparty`
- Generates grid with potential bird symbol connections
- **Identifies stage-cleared symbols** (does NOT remove them)
- **Checks for regular bird symbol connections** and RNG validation
- Handles free spin triggering
- **Sets cascading flag** if bird symbol connections exist
- Returns grid with stage-cleared symbol positions AND connection info

#### 2. Stage-Cleared Processing - `/process-stage-cleared/birdsparty`
- **Removes all stage-cleared symbols** from grid
- **Applies gravity** and fills with new symbols
- **Updates stage progress** count and checks for level advancement
- **Checks for NEW regular bird connections** in the updated grid after gravity
- **RNG validates new connections** and sets cascading flag
- Returns updated grid with potential new connections

#### 3. Cascade Phase - `/cascade/birdsparty`
- **Processes regular bird symbol connections** from previous steps
- **ENHANCED**: Detects stage-cleared symbols that appear after gravity
- Handles cascading mechanics (remove → gravity → find new connections)
- **Returns stage-cleared detection info** for client to process via stage-cleared endpoint
- Continues until no more connections exist
- Handles RNG integration for subsequent connections

## API Interaction Flow

### 1. Basic Spin with Stage-Cleared Symbols

#### Initial Spin Request
```json
POST /spin/birdsparty
{
  "client_id": "client_id_here",
  "game_id": "birdsparty",
  "player_id": "player_id_here",
  "bet_id": "bet_id_here",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [],
    "stageProgress": 5,
    "gameMode": "base",
    "freeSpins": { "remaining": 0, "totalAwarded": 0, "multiplier": 1.0 },
    "totalWin": 0,
    "cascading": false,
    "lastConnections": [],
    "cascadeCount": 0,
    "stageClearedSymbols": []
  }
}
```

#### Spin Response with Stage-Cleared Symbols
```json
{
  "status": "success",
  "message": "",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [
      ["orange_slice", "blue_owl", "yellow_owl", "green_owl"],
      ["red_owl", "purple_owl", "orange_slice", "blue_owl"],
      ["blue_owl", "green_owl", "purple_owl", "red_owl"],
      ["purple_owl", "yellow_owl", "blue_owl", "green_owl"]
    ],
    "stageProgress": 5,
    "gameMode": "base",
    "freeSpins": { "remaining": 0, "totalAwarded": 0, "multiplier": 1.0 },
    "totalWin": 0,
    "cascading": false,
    "lastConnections": [],
    "cascadeCount": 0,
    "stageClearedSymbols": [
      { "symbol": "orange_slice", "position": {"x": 0, "y": 0} },
      { "symbol": "orange_slice", "position": {"x": 2, "y": 1} }
    ]
  },
  "stageClearedSymbols": [
    { "symbol": "orange_slice", "position": {"x": 0, "y": 0} },
    { "symbol": "orange_slice", "position": {"x": 2, "y": 1} }
  ],
  "hasStageCleared": true,
  "totalCost": 0.1
}
```

### 2. Processing Stage-Cleared Symbols

#### Stage-Cleared Processing Request
```json
POST /process-stage-cleared/birdsparty
{
  "client_id": "client_id_here",
  "game_id": "birdsparty",
  "player_id": "player_id_here",
  "bet_id": "stage_cleared_001",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [
      ["orange_slice", "blue_owl", "yellow_owl", "green_owl"],
      ["red_owl", "purple_owl", "orange_slice", "blue_owl"],
      ["blue_owl", "green_owl", "purple_owl", "red_owl"],
      ["purple_owl", "yellow_owl", "blue_owl", "green_owl"]
    ],
    "stageProgress": 5,
    "gameMode": "base",
    "freeSpins": { "remaining": 0, "totalAwarded": 0, "multiplier": 1.0 },
    "totalWin": 0,
    "cascading": false,
    "lastConnections": [],
    "cascadeCount": 0,
    "stageClearedSymbols": [
      { "symbol": "orange_slice", "position": {"x": 0, "y": 0} },
      { "symbol": "orange_slice", "position": {"x": 2, "y": 1} }
    ]
  }
}
```

#### Stage-Cleared Processing Response
```json
{
  "status": "success",
  "message": "",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [
      ["green_owl", "blue_owl", "yellow_owl", "green_owl"],
      ["red_owl", "purple_owl", "yellow_owl", "blue_owl"],
      ["blue_owl", "green_owl", "purple_owl", "red_owl"],
      ["purple_owl", "yellow_owl", "blue_owl", "green_owl"]
    ],
    "stageProgress": 7,
    "gameMode": "base",
    "freeSpins": { "remaining": 0, "totalAwarded": 0, "multiplier": 1.0 },
    "totalWin": 0,
    "cascading": false,
    "lastConnections": [],
    "cascadeCount": 0,
    "stageClearedSymbols": []
  },
  "stageClearedCount": 2,
  "levelAdvanced": false,
  "totalCost": 0
}
```

### 3. Level Advancement During Stage-Cleared Processing

#### Level Advancement Response
```json
{
  "status": "success",
  "message": "",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 2,
    "gridSize": 5,
    "grid": [
      ["purple_owl", "blue_owl", "yellow_owl", "green_owl", "red_owl"],
      ["green_owl", "purple_owl", "yellow_owl", "blue_owl", "green_owl"],
      ["yellow_owl", "green_owl", "purple_owl", "red_owl", "blue_owl"],
      ["blue_owl", "yellow_owl", "blue_owl", "green_owl", "purple_owl"],
      ["red_owl", "blue_owl", "green_owl", "yellow_owl", "red_owl"]
    ],
    "stageProgress": 1,
    "gameMode": "base",
    "freeSpins": { "remaining": 0, "totalAwarded": 0, "multiplier": 1.0 },
    "totalWin": 0,
    "cascading": false,
    "lastConnections": [],
    "cascadeCount": 0,
    "stageClearedSymbols": []
  },
  "stageClearedCount": 4,
  "levelAdvanced": true,
  "oldLevel": 1,
  "newLevel": 2,
  "totalCost": 0
}
```

### 4. Enhanced Cascade Processing with Stage-Cleared Detection

#### Cascade Request
```json
POST /cascade/birdsparty
{
  "client_id": "client_id_here",
  "game_id": "birdsparty", 
  "player_id": "player_id_here",
  "bet_id": "cascade_001",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [
      ["green_owl", "blue_owl", "yellow_owl", "green_owl"],
      ["red_owl", "purple_owl", "yellow_owl", "blue_owl"],
      ["blue_owl", "green_owl", "purple_owl", "red_owl"],
      ["purple_owl", "yellow_owl", "blue_owl", "green_owl"]
    ],
    "stageProgress": 7,
    "gameMode": "base",
    "freeSpins": { "remaining": 0, "totalAwarded": 0, "multiplier": 1.0 },
    "totalWin": 0,
    "cascading": false,
    "lastConnections": [],
    "cascadeCount": 0,
    "stageClearedSymbols": []
  }
}
```

#### Enhanced Cascade Response with Stage-Cleared Detection
```json
{
  "status": "success",
  "message": "",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [
      ["green_owl", "blue_owl", "yellow_owl", "green_owl"],
      ["red_owl", "purple_owl", "yellow_owl", "blue_owl"],
      ["blue_owl", "green_owl", "purple_owl", "red_owl"],
      ["purple_owl", "yellow_owl", "blue_owl", "green_owl"]
    ],
    "stageProgress": 7,
    "gameMode": "base",
    "freeSpins": { "remaining": 0, "totalAwarded": 0, "multiplier": 1.0 },
    "totalWin": 0.20,
    "cascading": true,
    "lastConnections": [
      {
        "symbol": "green_owl",
        "positions": [
          {"x": 0, "y": 0}, {"x": 3, "y": 0}, {"x": 1, "y": 2}, {"x": 3, "y": 3}
        ],
        "count": 4,
        "payout": 0.20
      }
    ],
    "cascadeCount": 1
  },
  "connections": [
    {
      "symbol": "green_owl",
      "positions": [
        {"x": 0, "y": 0}, {"x": 3, "y": 0}, {"x": 1, "y": 2}, {"x": 3, "y": 3}
      ],
      "count": 4,
      "payout": 0.20
    }
  ],
  "stageClearedSymbols": [
    { "symbol": "orange_slice", "position": {"x": 1, "y": 0} }
  ],
  "hasStageCleared": true,
  "totalCost": 0
}
```
## Error Handling

### Stage-Cleared Processing Errors
```json
{
  "status": "error",
  "message": "Invalid grid dimensions for level 2"
}
```

### Missing Stage-Cleared Symbols
```json
{
  "status": "error", 
  "message": "No stage-cleared symbols found to process"
}
```

### Common Errors
- "Invalid bet amount" - Bet amount not in allowed values (0.1, 0.2, 0.3, 0.5, 1.0)
- "client_id is required" - Missing required field
- "Failed to retrieve game settings" - Settings service issue
- "Failed to determine outcome" - RNG service issue

## Testing and Debugging

### Debug Information
- Monitor server logs for stage-cleared symbol detection and removal
- Track level advancement timing and grid regeneration
- Verify stage progress accumulation (only stage-cleared symbols count)
- Check grid expansion and symbol generation per level
- **NEW**: Monitor cascade stage-cleared detection and processing flow

### Test Scenarios
1. **Stage-Cleared Symbol Priority**: Verify stage-cleared symbols are removed before connections
2. **Level Advancement**: Test progression through all 3 levels with grid expansion
3. **Symbol Separation**: Ensure bird symbols and stage-cleared symbols are handled separately
4. **Free Spins with Stage-Cleared**: Test free spins containing stage-cleared symbols
5. **Overflow Progress**: Test carrying excess progress to new levels
6. **Grid Regeneration**: Verify new grids after level advancement
7. **ENHANCED**: Cascade Stage-Cleared Detection: Test stage-cleared symbols appearing during cascades
8. **ENHANCED**: Complex Flow: Test cascade → stage-cleared → cascade sequences
9. **ENHANCED**: Level Advancement During Cascade: Test level advancement mid-cascade sequence

### Performance Considerations
- **Three-Endpoint Flow**: Ensure smooth transitions between endpoints
- **Grid Resizing**: Optimize UI transitions when changing grid sizes
- **Symbol Management**: Efficient loading/unloading of level-specific symbols
- **Animation Sequencing**: Coordinate stage-cleared removal with gravity effects
- **ENHANCED**: Cascade Detection: Optimize stage-cleared symbol detection during cascades
- **ENHANCED**: Complex Flow Management: Handle cascade → stage-cleared → cascade chains efficiently

### Debug Output Examples

#### Enhanced Cascade Flow with Stage-Cleared Detection
```
Starting enhanced cascade sequence
Processing cascade #1
Found 1 bird connections
Stage-cleared symbols detected during cascade: 2
ProcessStageCleared completed: stageClearedCount=2, levelAdvanced=false, cascading=true
Processing cascade #2
Found 0 bird connections
No more cascades, ending sequence
Enhanced cascade sequence completed
```

#### Level Advancement During Cascade
```
Starting enhanced cascade sequence
Processing cascade #1
Found 2 bird connections
Stage-cleared symbols detected during cascade: 3
ProcessStageCleared completed: stageClearedCount=3, levelAdvanced=true, oldLevel=1, newLevel=2
Level advanced from 1 to 2, grid: 5x5
Processing cascade #2
Found 1 bird connections
Enhanced cascade sequence completed
```

## Key Enhancements in This Version

### 1. Enhanced Cascade Handler
- **Stage-Cleared Detection**: Cascade handler now detects stage-cleared symbols that appear after gravity
- **Clean Separation**: Does not process stage-cleared symbols, only detects and reports them
- **Proper Flow Control**: Returns detection info for client to handle appropriately

### 2. Improved Client Flow
- **Complex Sequence Handling**: Handles cascade → stage-cleared → cascade chains seamlessly
- **Enhanced Debugging**: Comprehensive logging for flow tracking
- **Robust Error Handling**: Graceful handling of complex interaction sequences

### 3. Architectural Benefits
- **Maintainable Code**: Each endpoint has a single, clear responsibility
- **Flexible Integration**: Client can orchestrate complex flows as needed
- **RNG Integrity**: Proper RNG validation for each symbol type in its dedicated endpoint
- **Scalable Design**: Easy to add new features or modify individual components

## Advanced Flow Patterns

### Pattern 1: Simple Cascade without Stage-Cleared
```
Spin → Cascade (bird connections) → Cascade (more connections) → End
```

### Pattern 2: Stage-Cleared Priority Flow
```
Spin → Process-Stage-Cleared → Cascade (new connections) → End
```

### Pattern 3: Complex Mixed Flow
```
Spin → Process-Stage-Cleared → Cascade (connections + stage-cleared detected) → 
Process-Stage-Cleared → Cascade → End
```

### Pattern 4: Level Advancement Flow
```
Spin → Process-Stage-Cleared (level up) → Cascade (on new level) → 
Cascade (stage-cleared detected) → Process-Stage-Cleared → End
```

## Implementation Notes

### Backend Responsibilities
1. **Spin Handler**: Grid generation, initial detection, RNG validation
2. **Process-Stage-Cleared Handler**: Stage-cleared removal, level advancement, new grid analysis
3. **Cascade Handler**: Bird connection processing, stage-cleared detection (not processing)

### Client Responsibilities
1. **Flow Orchestration**: Coordinate between endpoints based on response flags
2. **Animation Management**: Handle visual transitions and timing
3. **State Management**: Maintain current game state throughout complex flows
4. **User Experience**: Provide clear feedback during complex sequences

### Critical Success Factors
1. **Proper Detection**: Ensure all stage-cleared symbols are detected when they appear
2. **Clean Separation**: Never mix bird connection processing with stage-cleared processing
3. **State Consistency**: Maintain accurate game state across all endpoint calls
4. **Error Recovery**: Handle network issues and invalid states gracefully

This enhanced implementation provides a robust, scalable foundation for the Birds Party game with proper separation of concerns and comprehensive stage-cleared symbol handling throughout all game phases!