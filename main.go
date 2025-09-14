package main

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	initialHP         = 5
	maxHP             = 10                     // Cap maximum HP
	initialMoveDelay  = 100 * time.Millisecond // Balanced movement speed
	initialEnemyDelay = 500 * time.Millisecond
	maxSpawnAttempts  = 100 // Prevent infinite loops
)

type position struct{ x, y int }

type tile int

const (
	empty tile = iota
	playerTile
	treasure
	trap
	enemy
	potion
)

type gameState int

const (
	menu gameState = iota
	paused
	playing
	gameOver
)

type model struct {
	state              gameState
	player             position
	playerHP           int
	treasures          []position
	traps              []position
	enemies            []position
	items              []position
	score              int
	level              int
	treasuresToCollect int
	board              [][]tile
	lastMove           time.Time
	moveDelay          time.Duration
	enemyDelay         time.Duration
	lastEnemyMove      time.Time
	pendingDirection   position
	width              int
	height             int
	countdown          int
	lastCountdown      time.Time
	countdownActive    bool
}

type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel() model {
	// Default size, will be updated on first resize
	width, height := 40, 20
	return model{
		state:      menu,
		playerHP:   initialHP,
		score:      0,
		level:      1,
		moveDelay:  initialMoveDelay,
		enemyDelay: initialEnemyDelay,
		width:      width,
		height:     height,
		board:      make([][]tile, height),
		countdown:  -1,
	}
}

func (m *model) resizeBoard(newWidth, newHeight int) {
	// Ensure minimum size
	if newWidth < 20 {
		newWidth = 20
	}
	if newHeight < 10 {
		newHeight = 10
	}

	m.width = newWidth
	m.height = newHeight

	// Recreate board with new dimensions
	m.board = make([][]tile, newHeight)
	for y := 0; y < newHeight; y++ {
		m.board[y] = make([]tile, newWidth)
	}

	// Adjust player position if out of bounds
	if m.player.x >= newWidth {
		m.player.x = newWidth - 1
	}
	if m.player.y >= newHeight {
		m.player.y = newHeight - 1
	}

	m.updateBoard()
}

func (m *model) startLevel() {
	m.clearEntities()
	numTreasures := 5 + (m.level-1)*2
	m.treasuresToCollect = numTreasures
	m.spawnTreasures(numTreasures)
	m.spawnTraps(3 + m.level)
	m.spawnEnemies(2 + m.level/2)
	// Spawn fewer potions - only every 3rd level gets a potion
	if m.level%3 == 0 {
		m.spawnItems(1)
	}
	m.player = position{m.width / 2, m.height / 2}
	// Reset pending direction to prevent immediate movement
	m.pendingDirection = position{0, 0}
	m.updateBoard()
}

func (m *model) clearEntities() {
	m.treasures = nil
	m.traps = nil
	m.enemies = nil
	m.items = nil
}

func (m *model) spawnTreasures(n int) {
	for i := 0; i < n; i++ {
		m.treasures = append(m.treasures, randomEmpty(m))
	}
}

func (m *model) spawnTraps(n int) {
	for i := 0; i < n; i++ {
		m.traps = append(m.traps, randomEmpty(m))
	}
}

func (m *model) spawnEnemies(n int) {
	for i := 0; i < n; i++ {
		m.enemies = append(m.enemies, randomEmpty(m))
	}
}

func (m *model) spawnItems(n int) {
	for i := 0; i < n; i++ {
		m.items = append(m.items, randomEmpty(m))
	}
}

func randomEmpty(m *model) position {
	occupied := map[position]bool{m.player: true}
	for _, p := range m.treasures {
		occupied[p] = true
	}
	for _, p := range m.traps {
		occupied[p] = true
	}
	for _, p := range m.enemies {
		occupied[p] = true
	}
	for _, p := range m.items {
		occupied[p] = true
	}

	for attempts := 0; attempts < maxSpawnAttempts; attempts++ {
		pos := position{rand.Intn(m.width), rand.Intn(m.height)}
		if !occupied[pos] {
			return pos
		}
	}

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			pos := position{x, y}
			if !occupied[pos] {
				return pos
			}
		}
	}
	panic("No empty tiles left!")
}

func contains(arr []position, p position) bool {
	for _, v := range arr {
		if v == p {
			return true
		}
	}
	return false
}

func containsPos(arr []position, p position, ignoreIdx int) bool {
	for j, v := range arr {
		if j != ignoreIdx && v == p {
			return true
		}
	}
	return false
}

func (m *model) updateBoard() {
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			m.board[y][x] = empty
		}
	}
	m.board[m.player.y][m.player.x] = playerTile
	for _, t := range m.treasures {
		m.board[t.y][t.x] = treasure
	}
	for _, t := range m.traps {
		m.board[t.y][t.x] = trap
	}
	for _, e := range m.enemies {
		m.board[e.y][e.x] = enemy
	}
	for _, i := range m.items {
		m.board[i.y][i.x] = potion
	}
}

func (m model) Init() tea.Cmd { return doTick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case menu:
			if msg.String() == "enter" || msg.String() == " " {
				m.state = playing
				m.startCountdown()
				return m, doTick()
			} else if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		case paused:
			if msg.String() == "p" {
				m.state = playing
				return m, doTick()
			} else if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		case playing:
			if msg.String() == "p" {
				m.state = paused
				return m, nil
			}
			// Don't process movement during countdown
			if !m.countdownActive {
				switch msg.String() {
				case "w", "up":
					m.pendingDirection = position{0, -1}
				case "s", "down":
					m.pendingDirection = position{0, 1}
				case "a", "left":
					m.pendingDirection = position{-1, 0}
				case "d", "right":
					m.pendingDirection = position{1, 0}
				case "q", "ctrl+c":
					return m, tea.Quit
				default:
					return m, nil
				}
			}
			return m, doTick()
		case gameOver:
			if msg.String() == "r" {
				return initialModel(), doTick()
			} else if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}

	case tickMsg:
		if m.state == playing {
			// Process countdown first
			if m.countdownActive {
				m.processCountdown()
			} else {
				// Only process game logic if countdown is not active
				m.moveEnemies()
				m.processPlayerMovement()
			}
		}
		return m, doTick()

	case tea.WindowSizeMsg:
		// Calculate board size based on terminal size
		// Reserve space for UI elements (borders, status, controls)
		// Emojis take up 2 character widths, so divide by 2 for proper sizing
		boardWidth := (msg.Width - 4) / 2 // Account for borders and emoji width
		boardHeight := msg.Height - 8     // Account for borders and UI text

		// Ensure we have a reasonable minimum size
		if boardWidth < 20 {
			boardWidth = 20
		}
		if boardHeight < 10 {
			boardHeight = 10
		}

		m.resizeBoard(boardWidth, boardHeight)
		return m, nil
	}
	return m, nil
}

func (m *model) processPlayerMovement() {
	// Only move if enough time has passed and there's a pending direction
	if time.Since(m.lastMove) >= m.moveDelay && (m.pendingDirection.x != 0 || m.pendingDirection.y != 0) {
		newPos := position{m.player.x + m.pendingDirection.x, m.player.y + m.pendingDirection.y}
		if newPos.x >= 0 && newPos.x < m.width && newPos.y >= 0 && newPos.y < m.height {
			m.player = newPos
			m.lastMove = time.Now()
			m.checkCollisions()
		}
	}
}

func (m *model) moveEnemies() {
	if time.Since(m.lastEnemyMove) < m.enemyDelay {
		return
	}
	for i, e := range m.enemies {
		dirX := 0
		if e.x < m.player.x {
			dirX = 1
		} else if e.x > m.player.x {
			dirX = -1
		}
		dirY := 0
		if e.y < m.player.y {
			dirY = 1
		} else if e.y > m.player.y {
			dirY = -1
		}
		newX, newY := e.x, e.y
		if dirX != 0 {
			newX += dirX
		} else if dirY != 0 {
			newY += dirY
		}
		newPos := position{newX, newY}
		if newPos != m.player && !contains(m.treasures, newPos) && !contains(m.traps, newPos) &&
			!contains(m.items, newPos) && !containsPos(m.enemies, newPos, i) {
			e.x, e.y = newX, newY
		}
		m.enemies[i] = e
	}
	m.lastEnemyMove = time.Now()
	m.updateBoard()
}

func (m *model) checkCollisions() {
	for i, t := range m.treasures {
		if t == m.player {
			m.score++
			m.treasures = append(m.treasures[:i], m.treasures[i+1:]...)
			break
		}
	}
	for i, item := range m.items {
		if item == m.player {
			// Only heal if below max HP
			if m.playerHP < maxHP {
				m.playerHP++
			}
			m.items = append(m.items[:i], m.items[i+1:]...)
			// Don't respawn potion immediately - let it respawn naturally
			break
		}
	}
	for i, t := range m.traps {
		if t == m.player {
			m.playerHP--
			m.traps = append(m.traps[:i], m.traps[i+1:]...)
			m.traps = append(m.traps, randomEmpty(m))
			if m.playerHP <= 0 {
				m.state = gameOver
			}
			break
		}
	}
	for i, e := range m.enemies {
		if e == m.player {
			m.playerHP--
			m.enemies[i] = randomEmpty(m)
			if m.playerHP <= 0 {
				m.state = gameOver
			}
			break
		}
	}
	if len(m.treasures) == 0 && m.treasuresToCollect > 0 {
		m.treasuresToCollect = 0
		m.level++
		m.enemyDelay = time.Duration(max(100, 500-m.level*30)) * time.Millisecond
		m.startCountdown()
	}
	m.updateBoard()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *model) getCountdownASCII(count int) string {
	switch count {
	case 3:
		return `                       ██████╗                      
                       ╚════██╗                     
                        █████╔╝                     
                        ╚═══██╗                     
                       ██████╔╝                     
                       ╚═════╝                      `
	case 2:
		return `                       ██████╗                      
                       ╚════██╗                     
                        █████╔╝                     
                       ██╔═══╝                      
                       ███████╗                     
                       ╚══════╝                     `
	case 1:
		return `                         ██╗                        
                         ██║                        
                         ██║                        
                         ██║                        
                         ██║                        
                         ╚═╝                        `
	case 0:
		return `   ███████╗████████╗ █████╗ ██████╗ ████████╗██╗    
   ██╔════╝╚══██╔══╝██╔══██╗██╔══██╗╚══██╔══╝██║    
   ███████╗   ██║   ███████║██████╔╝   ██║   ██║    
   ╚════██║   ██║   ██╔══██║██╔══██╗   ██║   ╚═╝    
   ███████║   ██║   ██║  ██║██║  ██║   ██║   ██╗    
   ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚═╝    `
	default:
		return ""
	}
}

func (m *model) startCountdown() {
	m.countdown = 3
	m.countdownActive = true
	m.lastCountdown = time.Now()
	// Load the level first so we can show it with the overlay
	m.startLevel()
}

func (m *model) processCountdown() {
	if time.Since(m.lastCountdown) >= time.Second {
		m.countdown--
		m.lastCountdown = time.Now()

		if m.countdown < 0 {
			// Countdown finished, activate gameplay
			m.countdownActive = false
		}
	}
}

func renderTile(t tile) string {
	switch t {
	case empty:
		return "⬛"
	case playerTile:
		return "🧙"
	case treasure:
		return "💰"
	case trap:
		return "⚠️"
	case enemy:
		return "👹"
	case potion:
		return "🧪"
	}
	return "?"
}

func (m model) View() string {
	var s string
	switch m.state {
	case menu:
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")).Align(lipgloss.Center).Width(m.width)
		instruct := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Align(lipgloss.Center).Width(m.width)
		s += title.Render("🛡️ DUNGEON DASH") + "\n\n"
		s += instruct.Render("Use WASD or arrow keys to move") + "\n"
		s += instruct.Render("Collect all treasures 💰 per level, avoid traps ⚠️ & enemies 👹") + "\n"
		s += instruct.Render("🧪 potions heal | Levels get harder!") + "\n\n"
		s += instruct.Render("Press ENTER to start or Q to quit") + "\n"
	case paused:
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")).Align(lipgloss.Center).Width(m.width)
		s += title.Render("PAUSED") + "\n\n"
		s += "Press P to resume or Q to quit\n"
	case playing, gameOver:
		board := ""
		for y := 0; y < m.height; y++ {
			for x := 0; x < m.width; x++ {
				board += renderTile(m.board[y][x])
			}
			board += "\n"
		}
		border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
		boardWithBorder := border.Render(board)

		// Apply countdown overlay using lipgloss.Place (like the example's dialog)
		if m.countdownActive {
			var countdownText string
			if m.countdown > 0 {
				countdownText = m.getCountdownASCII(m.countdown)
			} else {
				countdownText = m.getCountdownASCII(0) // "START!"
			}

			// Create a dialog box style like in the example
			dialogBoxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#874BFD")).
				Padding(1, 2).
				Background(lipgloss.Color("#2D3748")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Align(lipgloss.Center)

			dialog := dialogBoxStyle.Render(countdownText)

			// Use lipgloss.Place to overlay the dialog on top of the board
			// Use the exact board dimensions to prevent layout shift
			boardWithBorder = lipgloss.Place(
				lipgloss.Width(boardWithBorder),
				lipgloss.Height(boardWithBorder),
				lipgloss.Center,
				lipgloss.Center,
				dialog,
				lipgloss.WithWhitespaceChars("⬛"),
				lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}),
			)
		}

		s += boardWithBorder
		s += fmt.Sprintf("\nScore: %d | HP: %d | Level: %d | Treasures Left: %d\n", m.score, m.playerHP, m.level, len(m.treasures))
		if m.state == gameOver {
			over := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Align(lipgloss.Center).Width(m.width)
			s += over.Render("GAME OVER! Press R to restart or Q to quit\n")
		} else {
			if m.countdownActive {
				s += "Get ready to move!\n"
			} else {
				s += "Controls: WASD/Arrows: Move | P: Pause | Q: Quit\n"
			}
		}
	}
	return s
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
