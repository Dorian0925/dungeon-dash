package main

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	width             = 40
	height            = 20
	initialHP         = 5
	initialMoveDelay  = 200 * time.Millisecond
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
	board              [height][width]tile
	lastMove           time.Time
	moveDelay          time.Duration
	enemyDelay         time.Duration
	lastEnemyMove      time.Time
}

type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel() model {
	return model{
		state:      menu,
		playerHP:   initialHP,
		score:      0,
		level:      1,
		moveDelay:  initialMoveDelay,
		enemyDelay: initialEnemyDelay,
	}
}

func (m *model) startLevel() {
	m.clearEntities()
	numTreasures := 5 + (m.level-1)*2
	m.treasuresToCollect = numTreasures
	m.spawnTreasures(numTreasures)
	m.spawnTraps(3 + m.level)
	m.spawnEnemies(2 + m.level/2)
	m.spawnItems(1 + m.level/2)
	m.player = position{width / 2, height / 2}
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
		pos := position{rand.Intn(width), rand.Intn(height)}
		if !occupied[pos] {
			return pos
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
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
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
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
				m.startLevel()
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
			dir := position{0, 0}
			switch msg.String() {
			case "w", "up":
				dir = position{0, -1}
			case "s", "down":
				dir = position{0, 1}
			case "a", "left":
				dir = position{-1, 0}
			case "d", "right":
				dir = position{1, 0}
			case "q", "ctrl+c":
				return m, tea.Quit
			default:
				return m, nil
			}
			if time.Since(m.lastMove) >= m.moveDelay {
				newPos := position{m.player.x + dir.x, m.player.y + dir.y}
				if newPos.x >= 0 && newPos.x < width && newPos.y >= 0 && newPos.y < height {
					m.player = newPos
					m.lastMove = time.Now()
					m.checkCollisions()
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
			m.moveEnemies()
		}
		return m, doTick()
	}
	return m, nil
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
			m.playerHP++
			m.items = append(m.items[:i], m.items[i+1:]...)
			m.items = append(m.items, randomEmpty(m))
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
			m.state = gameOver
			m.enemies[i] = randomEmpty(m)
			break
		}
	}
	if len(m.treasures) == 0 && m.treasuresToCollect > 0 {
		m.treasuresToCollect = 0
		m.level++
		m.enemyDelay = time.Duration(max(100, 500-m.level*30)) * time.Millisecond
		m.startLevel()
	}
	m.updateBoard()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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
		return "⚠️ "
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
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")).Align(lipgloss.Center).Width(width)
		instruct := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Align(lipgloss.Center).Width(width)
		s += title.Render("🛡️ DUNGEON DASH") + "\n\n"
		s += instruct.Render("Use WASD or arrow keys to move") + "\n"
		s += instruct.Render("Collect all treasures 💰 per level, avoid traps ⚠️ & enemies 👹") + "\n"
		s += instruct.Render("🧪 potions heal | Levels get harder!") + "\n\n"
		s += instruct.Render("Press ENTER to start or Q to quit") + "\n"
	case paused:
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")).Align(lipgloss.Center).Width(width)
		s += title.Render("PAUSED") + "\n\n"
		s += "Press P to resume or Q to quit\n"
	case playing, gameOver:
		board := ""
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				board += renderTile(m.board[y][x])
			}
			board += "\n"
		}
		border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
		s += border.Render(board)
		s += fmt.Sprintf("\nScore: %d | HP: %d | Level: %d | Treasures Left: %d\n", m.score, m.playerHP, m.level, len(m.treasures))
		if m.state == gameOver {
			over := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Align(lipgloss.Center).Width(width)
			s += over.Render("GAME OVER! Press R to restart or Q to quit\n")
		} else {
			s += "Controls: WASD/Arrows: Move | P: Pause | Q: Quit\n"
		}
	}
	return s
}

func main() {
	rand.Seed(time.Now().UnixNano())
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
