package game

import "terminal-td/internal/entities"

type TowerTemplate struct {
	Type        entities.TowerType
	Name        string
	Cost        int
	Range       float64
	Damage      float64
	FireRate    float64
	Symbol      rune
	Color       int
	Description string
}

func GetTowerTemplates() map[entities.TowerType]TowerTemplate {
	return map[entities.TowerType]TowerTemplate{
		entities.TowerBasic: {
			Type:        entities.TowerBasic,
			Name:        "Basic Tower",
			Cost:        50,
			Range:       5.0,
			Damage:      10.0,
			FireRate:    1.0,
			Symbol:      'T',
			Color:       3,
			Description: "Standard tower with balanced stats",
		},
	}
}
