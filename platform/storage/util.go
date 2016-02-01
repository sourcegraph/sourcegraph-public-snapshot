package storage

import "encoding/json"

// PutJSON marshals and puts the JSON form of v into the storage system at
// <bucket>/<key>. Example usage:
//
// 	v := &Person{
// 		Name: "Billy",
// 		Age:  16,
// 	}
// 	if err := storage.PutJSON(sys, "people", "billy.json", v); err != nil {
// 		panic(err)
// 	}
//
func PutJSON(sys System, bucket, key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return sys.Put(bucket, key, data)
}

// GetJSON gets and unmarshals <bucket>/<key> from te storage system as JSON
// into v. Example usage:
//
// 	v := &Person{}
// 	if err := storage.GetJSON(sys, "people", "billy.json", &v); err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(v.Name) // "Billy"
// 	fmt.Println(v.Age)  // 16
//
func GetJSON(sys System, bucket, key string, v interface{}) error {
	data, err := sys.Get(bucket, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
