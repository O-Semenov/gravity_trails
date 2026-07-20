package game

import (
	"encoding/json"
	"os"
)

type SaveData struct {
	HighScore  float64            `json:"high_score"` // Для обратной совместимости
	HighScores map[string]float64 `json:"high_scores"`
}

// loadHighScore считывает рекорд из save.json
func (g *Game) loadHighScore() {
	g.allHighScores = make(map[string]float64)
	for _, m := range GameMaps {
		g.allHighScores[m.ID] = 0.0
	}

	file, err := os.Open("save.json")
	if err != nil {
		g.maxDistance = 0.0
		return
	}
	defer file.Close()

	var data SaveData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		g.maxDistance = 0.0
		return
	}

	// Миграция старого рекорда
	if data.HighScore > 0.0 {
		g.allHighScores["hills"] = data.HighScore
	}

	for id, score := range data.HighScores {
		g.allHighScores[id] = score
	}

	activeMap := GameMaps[g.currentMapIndex]
	g.maxDistance = g.allHighScores[activeMap.ID]
}

// saveHighScore сохраняет рекорд в save.json
func (g *Game) saveHighScore() {
	activeMap := GameMaps[g.currentMapIndex]
	if g.currentSessionMax > g.allHighScores[activeMap.ID] {
		g.allHighScores[activeMap.ID] = g.currentSessionMax
	}
	g.maxDistance = g.allHighScores[activeMap.ID]

	data := SaveData{
		HighScore:  g.allHighScores["hills"], // Совместимость
		HighScores: g.allHighScores,
	}

	file, err := os.Create("save.json")
	if err != nil {
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(data)
}
