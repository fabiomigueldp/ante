package engine

import "strings"

type BlindLevel struct {
	Level int
	SB    int
	BB    int
	Ante  int
}

type BlindStructure struct {
	Levels        []BlindLevel
	HandsPerLevel int
}

func TournamentBlinds(speed string) BlindStructure {
	levels := []BlindLevel{
		{Level: 1, SB: 1, BB: 2, Ante: 0},
		{Level: 2, SB: 2, BB: 4, Ante: 0},
		{Level: 3, SB: 3, BB: 6, Ante: 1},
		{Level: 4, SB: 5, BB: 10, Ante: 1},
		{Level: 5, SB: 7, BB: 15, Ante: 2},
		{Level: 6, SB: 10, BB: 20, Ante: 3},
		{Level: 7, SB: 15, BB: 30, Ante: 4},
		{Level: 8, SB: 20, BB: 40, Ante: 5},
		{Level: 9, SB: 30, BB: 60, Ante: 8},
		{Level: 10, SB: 50, BB: 100, Ante: 10},
		{Level: 11, SB: 75, BB: 150, Ante: 15},
		{Level: 12, SB: 100, BB: 200, Ante: 25},
	}
	handsPerLevel := 12
	switch strings.ToLower(speed) {
	case "turbo":
		handsPerLevel = 6
	case "slow":
		handsPerLevel = 20
	}
	return BlindStructure{Levels: levels, HandsPerLevel: handsPerLevel}
}

func HeadsUpBlinds() BlindStructure {
	return BlindStructure{
		HandsPerLevel: 15,
		Levels: []BlindLevel{
			{Level: 1, SB: 1, BB: 2, Ante: 0},
			{Level: 2, SB: 2, BB: 4, Ante: 0},
			{Level: 3, SB: 3, BB: 6, Ante: 1},
			{Level: 4, SB: 5, BB: 10, Ante: 1},
			{Level: 5, SB: 8, BB: 15, Ante: 2},
			{Level: 6, SB: 10, BB: 20, Ante: 3},
			{Level: 7, SB: 15, BB: 30, Ante: 5},
			{Level: 8, SB: 25, BB: 50, Ante: 5},
		},
	}
}

func CashGameBlinds(sb, bb int) BlindStructure {
	return BlindStructure{
		HandsPerLevel: 0,
		Levels:        []BlindLevel{{Level: 1, SB: sb, BB: bb, Ante: 0}},
	}
}
