package tui

import (
	"strings"
	"testing"

	"github.com/fabiomigueldp/ante/internal/engine"
)

func TestRenderHoleCardsDoesNotRenderZeroValueCards(t *testing.T) {
	rendered := RenderHoleCards([2]engine.Card{}, true)
	if strings.Contains(rendered, "0♠") {
		t.Fatalf("expected zero-value hole cards to avoid 0♠ rendering, got %q", rendered)
	}
	if !strings.Contains(rendered, CardBack()) {
		t.Fatalf("expected invalid visible hole cards to fall back to card backs, got %q", rendered)
	}
}

func TestRenderBigCardUsesPlaceholderForInvalidCard(t *testing.T) {
	rendered := RenderBigCard(engine.Card{})
	if strings.Contains(rendered, "0♠") {
		t.Fatalf("expected invalid big card render to avoid 0♠, got %q", rendered)
	}
	if !strings.Contains(rendered, "┌──┐") {
		t.Fatalf("expected placeholder card box for invalid card, got %q", rendered)
	}
}
