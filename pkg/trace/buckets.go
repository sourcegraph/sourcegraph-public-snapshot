package trace

// UserLatencyBuckets is a recommended list of buckets for use in prometheus
// histograms when measuring latency to users.
// Motivation: longer than 30s we don't care about. 2 is a general SLA we
// have. Otherwise rest is somewhat evenly spreadout to get good data
var UserLatencyBuckets = []float64{0.2, 0.5, 1, 2, 5, 10, 30}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_917(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
