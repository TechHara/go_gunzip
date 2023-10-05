package main

const WindowSize = MAX_DISTANCE * 3

type SlidingWindow struct {
	Data     []uint8
	Boundary int
}

func NewSlidingWindow() *SlidingWindow {
	return &SlidingWindow{make([]uint8, WindowSize), 0}
}

func (w *SlidingWindow) WriteBuffer() []uint8 {
	return w.Data[w.Boundary:]
}

func (w *SlidingWindow) Slide(n int) {
	end := w.Boundary + n
	if end > MAX_DISTANCE {
		delta := end - MAX_DISTANCE
		copy(w.Data, w.Data[delta:end])
		w.Boundary = MAX_DISTANCE
	} else {
		w.Boundary = end
	}
}
