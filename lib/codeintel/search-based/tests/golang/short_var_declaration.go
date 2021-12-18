package golang

import (
	"fmt"
)

func shortVarDeclaration() {
	message := "Hello!"
	message = "Hello!"
	message1, message2 := "a", "b"
	message1, message2 = message, message
	fmt.Println(message, message1, message2)
}
