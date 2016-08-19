package lightstep

import "github.com/opentracing/basictracer-go"

const defaultMaxSpans = 1000

type spansBuffer struct {
	rawSpans      []basictracer.RawSpan
	maxBufferSize int
}

func (b *spansBuffer) setDefaults() {
	b.maxBufferSize = defaultMaxSpans
	b.rawSpans = make([]basictracer.RawSpan, 0, b.maxBufferSize)
}

func (b *spansBuffer) setMaxBufferSize(size int) {
	b.maxBufferSize = size
}

func (b *spansBuffer) len() int {
	return len(b.rawSpans)
}

func (b *spansBuffer) cap() int {
	return b.maxBufferSize
}

func (b *spansBuffer) reset() {
	// Reuse the existing buffer if it's the correct size
	if cap(b.rawSpans) == b.maxBufferSize {
		b.rawSpans = b.rawSpans[:0]
	} else {
		b.rawSpans = make([]basictracer.RawSpan, 0, b.maxBufferSize)
	}
}

func (b *spansBuffer) current() []basictracer.RawSpan {
	dst := make([]basictracer.RawSpan, len(b.rawSpans))
	copy(dst, b.rawSpans)
	return dst
}

// addSpans returns the number of spans dropped (0 if all were added to the
// buffer).
func (b *spansBuffer) addSpans(spans []basictracer.RawSpan) (droppedSpans int) {
	space := b.maxBufferSize - len(b.rawSpans)
	count := space
	if len(spans) < count {
		count = len(spans)
	}
	if count > 0 {
		b.rawSpans = append(b.rawSpans, spans[:count]...)
	}
	droppedSpans = len(spans) - count
	return
}
