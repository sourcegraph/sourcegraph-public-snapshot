package core

func Join[A any, B any](fa func() (A, error), fb func() (B, error)) (A, B, error) {
	ca := make(chan A)
	cb := make(chan B)
	cerrA := make(chan error)
	cerrB := make(chan error)
	go func() {
		a, err := fa()
		if err != nil {
			cerrA <- err
			return
		}
		ca <- a
	}()
	go func() {
		b, err := fb()
		if err != nil {
			cerrB <- err
			return
		}
		cb <- b
	}()
	select {
	// TODO: combine potential multiple errors?
	case err := <-cerrA:
		return *new(A), *new(B), err
	case err := <-cerrB:
		return *new(A), *new(B), err
	case a := <-ca:
		select {
		case b := <-cb:
			return a, b, nil
		case err := <-cerrB:
			return *new(A), *new(B), err
		}
	case b := <-cb:
		select {
		case a := <-ca:
			return a, b, nil
		case err := <-cerrA:
			return *new(A), *new(B), err
		}
	}
}

func Race[A any, B any](fa func(stop chan struct{}) (A, error), fb func(stop chan struct{}) (B, error)) (A, B, bool, error) {
	ca := make(chan A)
	cb := make(chan B)
	cerr := make(chan error)
	stop := make(chan struct{})

	go func() {
		a, err := fa(stop)
		if err != nil {
			cerr <- err
			return
		}
		ca <- a
		close(ca)
	}()
	go func() {
		b, err := fb(stop)
		if err != nil {
			cerr <- err
			return
		}
		cb <- b
		close(cb)
	}()

	select {
	case a := <-ca:
		stop <- struct{}{}
		return a, *new(B), true, nil
	case b := <-cb:
		stop <- struct{}{}
		return *new(A), b, false, nil
	case <-cerr:
		select {
		case a := <-ca:
			return a, *new(B), true, nil
		case b := <-cb:
			return *new(A), b, false, nil
		case err := <-cerr:
			return *new(A), *new(B), false, err
		}
	}
}
