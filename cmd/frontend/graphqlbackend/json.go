package graphqlbackend

import "encoding/json"

// jsonValue implements the JSONValue scalar type. In GraphQL queries, it is represented the JSON
// representation of its Go value.
type jsonValue struct{ value interface{} }

func (jsonValue) ImplementsGraphQLType(name string) bool {
	return name == "JSONValue"
}

func (v *jsonValue) UnmarshalGraphQL(input interface{}) error {
	*v = jsonValue{value: input}
	return nil
}

func (v jsonValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_158(size int) error {
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
