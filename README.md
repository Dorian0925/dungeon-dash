# Dungeon Dash  

A terminal-based dungeon crawler game built with Go and Bubble Tea. Navigate through levels, collect treasures, avoid traps and enemies, and survive as long as possible.  

<p align="center">
  <img src="dungeon-dash.jpg" alt="Dungeon Dash gameplay" width="700">
</p>

## Features  

- Progressive difficulty with increasing levels  
- Real-time enemy AI that tracks player movement  
- Health system with healing potions  
- Trap avoidance mechanics  
- Score tracking and level progression  
- Responsive terminal UI with Unicode characters  

## Installation  

### Prerequisites  

- Go 1.25.0 or higher  

### Build from Source  

```bash  
git clone https://github.com/Cod-e-Codes/dungeon-dash.git  
cd dungeon-dash  
go mod download  
go build -o dungeon-dash
```

### Run directly with Go

```bash
go run main.go
```

## How to Play

### Objective

Collect all treasures on each level while avoiding traps and enemies. Each level increases in difficulty with more enemies and faster movement.

### Controls

- **Movement**: WASD or Arrow keys
- **Pause**: P
- **Quit**: Q or Ctrl+C
- **Restart**: R (when game over)

### Game Elements

- **Player**: Wizard character you control
- **Treasures**: Collect all to advance to next level
- **Traps**: Damage player on contact, respawn after being triggered
- **Enemies**: Move toward player, cause game over on contact
- **Potions**: Restore health when collected, respawn after use

### Mechanics

- Player starts with 5 HP
- Taking damage from traps reduces HP by 1
- Contact with enemies ends the game
- Potions increase HP by 1
- Enemy movement speed increases with each level
- More enemies and traps spawn as levels progress

## Technical Details

### Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) v1.3.9 - Terminal UI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) v1.1.0 - Style and layout library

### Game Configuration

- Board size: 40x20 characters
- Initial move delay: 200ms
- Initial enemy delay: 500ms (decreases with level)
- Starting HP: 5
- Treasures per level: 5 + (level-1) * 2
- Traps per level: 3 + level
- Enemies per level: 2 + level/2

## License

This project is licensed under the MIT License.

## Contributing

Contributions are welcome. Please submit issues and pull requests through GitHub.
