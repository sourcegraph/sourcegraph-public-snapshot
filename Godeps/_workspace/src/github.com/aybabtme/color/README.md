# Package `color`

Colorize your terminal strings.

```go
fmt.Printf("I'm in a %s world!\n", brush.Blue("blue"))
```

# Usage

Default `Brush` are available for your convenience.  You can invoke them directly

```go
fmt.Printf("This is %s\n", brush.Red("red"))
```

... or you can create new ones!
```go
weird := color.NewBrush(color.PurplePaint, color.CyanPaint)
fmt.Printf("This color is %s\n", weird("weird"))
```

Create a `Style`, which has some convenience methods :
```go
redBg := color.NewStyle(color.RedPaint, color.YellowPaint)
```

`Style.WithForeground` or `WithBackground` returns a new `Style`, with the applied
`Paint`.  Styles are immutable so the original one is left unchanged

```go
greenFg := redBg.WithForeground(color.GreenPaint)

// Style.Brush gives you a Brush that you can invoke directly to colorize strings.
green := greenFg.Brush()
fmt.Printf("This is %s but not really\n", green("kind of green"))
```

You can use it with all sorts of things :
```go
sout := log.New(os.Stdout, "["+brush.Green("OK").String()+"]\t", log.LstdFlags)
serr := log.New(os.Stderr, "["+brush.Red("OMG").String()+"]\t", log.LstdFlags)

sout.Printf("Everything was going %s until...", brush.Cyan("fine"))
serr.Printf("%s killed %s !!!", brush.Red("Locke"), brush.Blue("Jacob"))
```

That's it!

# Demo

![A coloured terminal](https://s3-us-west-2.amazonaws.com/aybabtme/color_demo.png "A fine terminal")

# Docs

[GoDoc!](http://godoc.org/github.com/aybabtme/color) (↫ this is a link)

# FAQ

> Does it work on Windows?

NO!

> It's spelled "colour"

NO!

> You're canadian, spell it "colour"

NO!
