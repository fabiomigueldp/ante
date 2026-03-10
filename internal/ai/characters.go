package ai

func Characters() []Character {
	return []Character{
		{ID: "riley_rookie", Profile: beginnerProfile("Riley Banks", "Rookie", "Wait, how much is the bet?", StyleStraight, 0.22, 0.08, 0.15, 0.02, 0.65, 0.00, 0.10)},
		{ID: "lucy_lucky", Profile: beginnerProfile("Lucy Malone", "Lucky", "I had a feeling about that card!", StyleStation, 0.42, 0.06, 0.18, 0.04, 0.85, 0.00, 0.20)},
		{ID: "greg_granite", Profile: intermediateProfile("Greg Holloway", "Granite", "Patience is a virtue. Folding is an art.", StyleNit, 0.12, 0.09, 0.18, 0.01, 0.22, 0.05, 0.02)},
		{ID: "eddie_steady", Profile: intermediateProfile("Eddie Park", "Steady", "I'll wait for the nuts, thanks.", StyleStraight, 0.18, 0.11, 0.22, 0.02, 0.35, 0.03, 0.03)},
		{ID: "rosa_river", Profile: intermediateProfile("Rosa Chen", "The River", "I'm not folding, I've already put too much in.", StyleStation, 0.40, 0.05, 0.20, 0.01, 0.92, 0.01, 0.05)},
		{ID: "alejandro_ace", Profile: intermediateProfile("Alejandro Vega", "Ace", "Keep it simple. Play good cards.", StyleTAG, 0.22, 0.16, 0.35, 0.03, 0.38, 0.02, 0.04)},
		{ID: "bruno_bulldog", Profile: intermediateProfile("Bruno Santos", "Bulldog", "Go big or go home.", StyleStraight, 0.20, 0.12, 0.46, 0.00, 0.18, 0.01, 0.06)},
		{ID: "sasha_smooth", Profile: intermediateProfile("Sasha Morin", "Smooth", "Let's not get ahead of ourselves.", StyleStraight, 0.28, 0.09, 0.18, 0.01, 0.42, 0.02, 0.02)},
		{ID: "tommy_sheriff", Profile: intermediateProfile("Tommy Kwon", "The Sheriff", "I don't believe you.", StyleHeroCaller, 0.26, 0.11, 0.28, 0.02, 0.68, 0.03, 0.05)},
		{ID: "jake_blitz", Profile: advancedProfile("Jake Beckett", "Blitz", "Why wait when you can raise?", StyleLAG, 0.35, 0.28, 0.82, 0.18, 0.34, 0.04, 0.10)},
		{ID: "carlos_maddog", Profile: advancedProfile("Carlos Rivera", "Mad Dog", "You think you can push ME around?!", StyleLAG, 0.32, 0.24, 0.76, 0.16, 0.36, 0.03, 0.92)},
		{ID: "phil_phantom", Profile: advancedProfile("Phil Okafor", "Phantom", "Check. ...Raise.", StyleTrappy, 0.24, 0.18, 0.58, 0.08, 0.32, 0.42, 0.06)},
		{ID: "carter_cowboy", Profile: advancedProfile("Carter Hayes", "Cowboy", "Yeehaw, let's see a flop.", StyleLAG, 0.34, 0.22, 0.62, 0.14, 0.30, 0.08, 0.08)},
		{ID: "wendy_wildcard", Profile: advancedProfile("Wendy Torres", "Wildcard", "All in. Again.", StyleManiac, 0.46, 0.31, 0.95, 0.24, 0.42, 0.00, 0.18)},
		{ID: "nina_grinder", Profile: advancedProfile("Nina Garcia", "Grinder", "Small pots, small risk, steady profit.", StyleTAG, 0.27, 0.21, 0.44, 0.06, 0.28, 0.09, 0.02)},
		{ID: "viktor_professor", Profile: expertProfile("Viktor Stein", "The Professor", "Every bet tells a story.", StyleBalanced, 0.25, 0.19, 0.57, 0.10, 0.24, 0.14, 0.03)},
		{ID: "diana_duchess", Profile: expertProfile("Diana Ashworth", "Duchess", "Darling, you really should have folded.", StyleBalanced, 0.30, 0.24, 0.68, 0.16, 0.22, 0.10, 0.04)},
		{ID: "ingrid_ice", Profile: expertProfile("Ingrid Volkov", "Ice", "...", StyleBalanced, 0.24, 0.20, 0.55, 0.09, 0.18, 0.16, 0.01)},
	}
}

func beginnerProfile(name, nickname, flavor string, style Style, vpip, pfr, agg, bluff, callDown, trap, tilt float64) Profile {
	return Profile{Name: name, Nickname: nickname, Flavor: flavor, Style: style, Skill: SkillBeginner, VPIP: vpip, PFR: pfr, Aggression: agg, Bluff: bluff, CallDown: callDown, Trap: trap, Tilt: tilt, ThinkMinMS: 350, ThinkMaxMS: 1800, PositionBias: 0.05, DrawBias: 0.10, LargeBetBias: 0.08, HeroCallBias: callDown, ThreeBetBias: 0.02}
}

func intermediateProfile(name, nickname, flavor string, style Style, vpip, pfr, agg, bluff, callDown, trap, tilt float64) Profile {
	return Profile{Name: name, Nickname: nickname, Flavor: flavor, Style: style, Skill: SkillIntermediate, VPIP: vpip, PFR: pfr, Aggression: agg, Bluff: bluff, CallDown: callDown, Trap: trap, Tilt: tilt, ThinkMinMS: 500, ThinkMaxMS: 2200, PositionBias: 0.12, DrawBias: 0.18, LargeBetBias: 0.16, HeroCallBias: callDown, ThreeBetBias: 0.08}
}

func advancedProfile(name, nickname, flavor string, style Style, vpip, pfr, agg, bluff, callDown, trap, tilt float64) Profile {
	return Profile{Name: name, Nickname: nickname, Flavor: flavor, Style: style, Skill: SkillAdvanced, VPIP: vpip, PFR: pfr, Aggression: agg, Bluff: bluff, CallDown: callDown, Trap: trap, Tilt: tilt, ThinkMinMS: 650, ThinkMaxMS: 2800, PositionBias: 0.20, DrawBias: 0.24, LargeBetBias: 0.24, HeroCallBias: callDown, ThreeBetBias: 0.16}
}

func expertProfile(name, nickname, flavor string, style Style, vpip, pfr, agg, bluff, callDown, trap, tilt float64) Profile {
	return Profile{Name: name, Nickname: nickname, Flavor: flavor, Style: style, Skill: SkillExpert, VPIP: vpip, PFR: pfr, Aggression: agg, Bluff: bluff, CallDown: callDown, Trap: trap, Tilt: tilt, ThinkMinMS: 900, ThinkMaxMS: 3600, PositionBias: 0.28, DrawBias: 0.28, LargeBetBias: 0.20, HeroCallBias: callDown, ThreeBetBias: 0.22}
}
