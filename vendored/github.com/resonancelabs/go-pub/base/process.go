package base

var processStartTimeMicros Micros

func init() {
	processStartTimeMicros = NowMicros()
}

func ProcessStartTimeMicros() Micros {
	return processStartTimeMicros
}
