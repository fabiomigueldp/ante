package ai

import (
	"math/rand"
	"sort"
)

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

func SelectCharacters(difficulty Difficulty, seats int, seed int64) []Character {
	pool := Characters()
	rng := rand.New(rand.NewSource(seed))
	allowed := make([]Character, 0, len(pool))
	for _, character := range pool {
		skill := character.Profile.Skill
		switch difficulty {
		case DifficultyEasy:
			if skill == SkillExpert || skill == SkillAdvanced {
				continue
			}
		case DifficultyMedium:
			if skill == SkillExpert && rng.Float64() < 0.5 {
				continue
			}
		case DifficultyHard:
			if skill == SkillBeginner {
				continue
			}
		}
		allowed = append(allowed, character)
	}
	rng.Shuffle(len(allowed), func(i, j int) { allowed[i], allowed[j] = allowed[j], allowed[i] })
	if seats > len(allowed) {
		seats = len(allowed)
	}
	selected := append([]Character(nil), allowed[:seats]...)
	sort.Slice(selected, func(i, j int) bool { return selected[i].Profile.Name < selected[j].Profile.Name })
	return selected
}
