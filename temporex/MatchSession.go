package main

import "time"

type MatchSession struct {
	sessionID string
	playerIDs []string
	createdAt time.Time
	groups    map[string][]string // groupID to playerIDs
}

func (ms *MatchSession) AddPlayer(playerID string) {
	ms.playerIDs = append(ms.playerIDs, playerID)
}

func (ms *MatchSession) AssignPlayerToGroup(playerID, groupID string) {
	if ms.groups == nil {
		ms.groups = make(map[string][]string)
	}
	ms.groups[groupID] = append(ms.groups[groupID], playerID)
}
