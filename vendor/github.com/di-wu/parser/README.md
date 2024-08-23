# Parser Building Blocks

This project contains a lexerless parser, it performs tokenization and parsing in a single step. This enables you to use
a language grammar that expresses both the lexical (word level), and the phrase level structure of the language.

##### Advantages

- Non-regular lexical structures are handled easier.
- Context insensitive.
- No token classification. This removes language reserved words such as 'for' in Go.
- Grammars are compositional (can be merged automatically).
- Only one metalanguage is needed. *

*I chose to use [PEGN](https://github.com/pegn) as metalanguage in the examples. This is **not** required, you can also
use other metalanguages like PEG or (A)BNF.*

##### Disadvantages

- ~~More complicated~~.
- Less efficient than lexer-parser with regard to both **time** and **memory**.

## Usage

The project two parts.

### Basic Parser

1. The main `parser` package with provides you with a basic parser with more control.

*Errors are ignored for simplicity!*

```go
p, _ := parser.New([]byte("raw data to parse"))
mark, err := p.Expect("two words")
```

The example above tries to parse the string `"two words"`. In case the parser **succeeded** in parsing the string, it
will return a mark to the last parsed rune. Otherwise, an error will be returned.

This way the implementer is responsible to capture any value if needed. The parser only will let you know if it
succeeded in parsing the given value.

##### Supported Values

- `rune` (`int` will get converted to runes for convenience).
- `string`.
- `AnonymousClass` (equal to `func(p *Parser) (*Cursor, bool)`).
- All operators defined in the `op` sub-package.

##### Customizing

The parser expects `UTF8` encoded strings by default. It is possible to use other decoders. This can be done by
implementing the `DecodeRune` callback. This is done in the [ELF example](./examples/elf).

It is also possible to provide additional supported operators or converters.

### AST Parser

2. The `ast` package which provides you an interface to immediately construct a syntax tree.

*Errors are ignored for simplicity!*

```go
p, _ := parser.New([]byte("raw data to parse"))
node, err := p.Expect(ast.Capture{
    TypeStrings: []string{"Digit"},
    Value: parser.CheckRuneFunc(func (r rune) bool {
        return '0' <= r && r <= '9'
    }),
})
```

This example tries to capture a single digit. In case the parser **succeeded** in parsing a digit, it will return a node
with the parsed digit as value. Otherwise, an error will be returned.

The AST parser in build on top of the basic parser. It is extended with a few more values.

##### Supported Values

- All values supported in the basic parser.
- `ParseNode` (equal to `func(p *Parser) (*Node, error)`)
- `Capture` (captures the value in a node)
- `LoopUp`

For more info check out the [documentation](https://pkg.go.dev/github.com/di-wu/parser), it contains examples and
descriptions for all functionality.

## Documentation

You can find the documentation [here](https://pkg.go.dev/github.com/di-wu/parser). Additional examples can be
found [here](./examples).

## Contributing

Contributions are welcome. Feel free to create a PR/issue for new features or bug fixes.

## License

Apache License found [here](LICENSE).
