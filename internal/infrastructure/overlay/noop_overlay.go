package overlay

// NoopOverlay is used when the client runs without RTSS integration.
type NoopOverlay struct{}

func NewNoopOverlay() *NoopOverlay {
	return &NoopOverlay{}
}

func (o *NoopOverlay) Init() {}

func (o *NoopOverlay) UpdateText(_ string) {}

func (o *NoopOverlay) Close() {}
