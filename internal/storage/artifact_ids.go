package storage

import (
	"fmt"
	"strings"
)

func TranscriptIDFromSessionID(sessionID string) (string, error) {
	base, err := baseTokenFromSessionID(sessionID)
	if err != nil {
		return "", err
	}
	return "trn_" + base, nil
}

func ChunkIDFromSessionID(sessionID string, chunkIndex int) (string, error) {
	base, err := baseTokenFromSessionID(sessionID)
	if err != nil {
		return "", err
	}
	if chunkIndex < 0 {
		return "", fmt.Errorf("chunk index must be >= 0")
	}
	return fmt.Sprintf("tch_%s_%06d", base, chunkIndex), nil
}

func CheckpointIDFromSessionID(sessionID string, handIndex int) (string, error) {
	base, err := baseTokenFromSessionID(sessionID)
	if err != nil {
		return "", err
	}
	if handIndex < 0 {
		return "", fmt.Errorf("hand index must be >= 0")
	}
	return fmt.Sprintf("ckp_%s_%06d", base, handIndex), nil
}

func SnapshotIDFromSessionID(sessionID string, handIndex int, seq uint64) (string, error) {
	base, err := baseTokenFromSessionID(sessionID)
	if err != nil {
		return "", err
	}
	if handIndex < 0 {
		return "", fmt.Errorf("hand index must be >= 0")
	}
	return fmt.Sprintf("snp_%s_%06d_%09d", base, handIndex, seq), nil
}

func baseTokenFromSessionID(sessionID string) (string, error) {
	const prefix = "ses_"
	if !strings.HasPrefix(sessionID, prefix) {
		return "", fmt.Errorf("invalid session id %q", sessionID)
	}
	base := strings.TrimPrefix(sessionID, prefix)
	if len(base) != 32 {
		return "", fmt.Errorf("invalid session id base token length for %q", sessionID)
	}
	for _, r := range base {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return "", fmt.Errorf("invalid session id base token %q", sessionID)
		}
	}
	return base, nil
}
