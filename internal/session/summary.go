package session

func (s *Session) persistSessionSummary() error {
	if s == nil || s.metrics == nil || s.transcript == nil {
		return nil
	}
	anchor, err := s.deps.TimeAnchorProvider.Now()
	if err != nil {
		return err
	}
	summary := s.metrics.BuildSummary(s, anchor, s.transcript.Head())
	if _, err := s.deps.ArtifactStore.SaveSessionSummaryArtifact(summary); err != nil {
		return err
	}
	s.Summary = &summary
	return nil
}
