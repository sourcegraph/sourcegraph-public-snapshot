package repos

func DefineNotify() {
	if notify == nil {
		notify = func(ch chan struct{}) {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
	}
}
