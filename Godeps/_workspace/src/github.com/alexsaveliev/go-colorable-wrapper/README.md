# go-colorable-wrapper

Wrapper over github.com/mattn/go-colorable that provides helper functions and variables

## Print methods
Mimic functions from standard `fmt` package but make them colorable-aware. Include
* `colorable.Print`
* `colorable.Printf`
* `colorable.Println`

## Colorable-aware streams
Provides `colorable.Stderr` and `colorable.Stdout` variables. They have the same capabilities as `os.Stderr` and `os.Stdout`, but colorable-aware

## ANSI codes
Provides functions to output colored text, change background color, or add some text effect such as bold or underline
