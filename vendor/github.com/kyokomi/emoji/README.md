# Emoji
Emoji is a simple golang package.

[![Build Status](https://drone.io/github.com/kyokomi/emoji/status.png)](https://drone.io/github.com/kyokomi/emoji/latest)
[![Coverage Status](https://coveralls.io/repos/kyokomi/emoji/badge.png?branch=master)](https://coveralls.io/r/kyokomi/emoji?branch=master)
[![GoDoc](https://godoc.org/github.com/kyokomi/emoji?status.svg)](https://godoc.org/github.com/kyokomi/emoji)

Get it:

```
go get gopkg.in/kyokomi/emoji.v1
```

Import it:

```
import (
	"gopkg.in/kyokomi/emoji.v1"
)
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/kyokomi/emoji"
)

func main() {
	fmt.Println("Hello Wolrd Emoji!")

	emoji.Println(":beer: Beer!!!")

	pizzaMessage := emoji.Sprint("I like a :pizza: and :sushi:!!")
	fmt.Println(pizzaMessage)
}
```

## Demo

![](https://raw.githubusercontent.com/kyokomi/emoji/master/screen/image.png)

## Reference

- [GitHub EMOJI CHEAT SHEET](http://www.emoji-cheat-sheet.com/)

## License

[MIT](https://github.com/kyokomi/emoji/blob/master/LICENSE)
