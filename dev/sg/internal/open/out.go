package open

import (
	"fmt"
)

func Prompt(req string) (string, error) {
	fmt.Print(req + " ")
	var answer string
	if _, err := fmt.Scan(&answer); err != nil {
		return "", err
	}
	return answer, nil
}
