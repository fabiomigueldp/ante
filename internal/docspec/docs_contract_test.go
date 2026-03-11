package docspec

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to locate current file")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("unable to locate repository root from %q: %v", root, err)
	}
	return root
}

func readDoc(t *testing.T, root, rel string) string {
	t.Helper()
	path := filepath.Join(root, rel)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("unable to read %s: %v", rel, err)
	}
	return string(data)
}

func containsAll(t *testing.T, docName, content string, required []string) {
	t.Helper()
	normalized := strings.ToLower(content)
	for _, term := range required {
		if !strings.Contains(normalized, strings.ToLower(term)) {
			t.Fatalf("%s is missing required text: %q", docName, term)
		}
	}
}

func TestRequiredDocsExist(t *testing.T) {
	root := repoRoot(t)
	required := []string{
		"docs/architecture.md",
		"docs/threat_model.md",
		"docs/glossary.md",
		"docs/decision_register.md",
		"docs/adr/ADR-001-artifact-store.md",
		"docs/adr/ADR-002-transcript-record-and-checkpointing.md",
		"docs/adr/ADR-003-session-authority-and-envelope.md",
		"docs/adr/ADR-004-gamevm-reducer-boundary.md",
		"docs/adr/ADR-005-time-anchor-provider.md",
		"docs/adr/ADR-006-canonical-encoding-and-stable-ids.md",
		"docs/adr/ADR-007-scope-guardrails.md",
	}

	for _, rel := range required {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("required doc missing: %s (%v)", rel, err)
		}
	}
}

func TestArchitectureDocContainsRequiredSections(t *testing.T) {
	root := repoRoot(t)
	content := readDoc(t, root, "docs/architecture.md")
	containsAll(t, "docs/architecture.md", content, []string{
		"project scope",
		"product boundaries",
		"non-goals",
		"explicit separation between sandbox, free-play multiplayer, and any future paid economy",
		"package map",
		"authoritative runtime data flow",
		"artifactstore responsibilities",
		"transcript and snapshot lifecycle",
		"timeanchorprovider contract",
		"free-table multiplayer layering",
		"balancechain overview",
		"failure domains",
		"migration expectations",
		"trust roots",
	})
}

func TestThreatModelContainsRequiredSections(t *testing.T) {
	root := repoRoot(t)
	content := readDoc(t, root, "docs/threat_model.md")
	containsAll(t, "docs/threat_model.md", content, []string{
		"assets and trust boundaries",
		"actor model",
		"storage attacker",
		"clock manipulator",
		"protocol attacker",
		"persistence attack surface",
		"transcript",
		"identity",
		"networking",
		"balance verification",
		"replay",
		"tampering",
		"impersonation",
		"equivocation",
		"clock-skew abuse",
		"lock-doubling",
		"bots are sandbox-only",
		"host-authoritative bringup harness risks",
		"incident response expectations",
		"unresolved risks",
	})
}

func TestGlossaryContainsMandatoryTerms(t *testing.T) {
	root := repoRoot(t)
	content := readDoc(t, root, "docs/glossary.md")
	containsAll(t, "docs/glossary.md", content, []string{
		"## artifactstore",
		"## transcriptrecord",
		"## session authority",
		"## gamevm",
		"## timeanchorprovider",
		"## canonical encoding",
		"## economy isolation",
		"## free table tier",
		"naming rules",
		"disallowed ambiguous terminology",
	})
}

func TestDecisionRegisterIndexesAcceptedAndReservedADRs(t *testing.T) {
	root := repoRoot(t)
	content := readDoc(t, root, "docs/decision_register.md")
	containsAll(t, "docs/decision_register.md", content, []string{
		"adr template definition",
		"status model",
		"index of accepted decisions",
		"reserved adr entries",
		"cross-reference rules",
		"adr-001",
		"adr-002",
		"adr-003",
		"adr-004",
		"adr-005",
		"adr-006",
		"adr-007",
		"adr-008",
		"adr-009",
		"adr-010",
		"adr-011",
	})
}

func TestADRFilesContainRequiredFields(t *testing.T) {
	root := repoRoot(t)
	adrFiles := []string{
		"docs/adr/ADR-001-artifact-store.md",
		"docs/adr/ADR-002-transcript-record-and-checkpointing.md",
		"docs/adr/ADR-003-session-authority-and-envelope.md",
		"docs/adr/ADR-004-gamevm-reducer-boundary.md",
		"docs/adr/ADR-005-time-anchor-provider.md",
		"docs/adr/ADR-006-canonical-encoding-and-stable-ids.md",
		"docs/adr/ADR-007-scope-guardrails.md",
	}
	required := []string{
		"status:",
		"date:",
		"## context",
		"## options considered",
		"## decision",
		"## consequences",
		"## migration impact",
		"## rollback notes",
		"## code paths impacted",
		"## tests impacted",
	}

	for _, rel := range adrFiles {
		content := readDoc(t, root, rel)
		containsAll(t, rel, content, required)
	}
}

func TestCoreVocabularyAppearsAcrossDocumentationSet(t *testing.T) {
	root := repoRoot(t)
	combined := strings.Join([]string{
		readDoc(t, root, "docs/architecture.md"),
		readDoc(t, root, "docs/threat_model.md"),
		readDoc(t, root, "docs/glossary.md"),
		readDoc(t, root, "docs/decision_register.md"),
	}, "\n")
	containsAll(t, "documentation set", combined, []string{
		"ArtifactStore",
		"TranscriptRecord",
		"Session Authority",
		"GameVM",
		"TimeAnchorProvider",
		"canonical encoding",
		"host-authoritative",
		"sandbox",
		"free-play",
		"paid economy",
	})
}
