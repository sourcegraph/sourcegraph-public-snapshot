# Media

## Logo

- Primary color: `#fc60a8`
- Secondary color: `#494368`
- Font: [`Orbitron`](https://fonts.google.com/specimen/Orbitron)
- Link: [Normal](#hello)
- Bold: **bold**
- Italic: *bold*
- Italic 2: _bold_
- Bold and italic: ***bold***
- Underlined: __underlined__
- Strikethrough: ~~strikethrough~~

Numbered lists

1. a
2. b
3. c

Numbered and nested lists
1. a
  - aa
    - aaa
  - aa2
2. b
  - bb
    - bbb
  - bb2
3. c
  - cc
    - ccc
  - cc2

You are free to use and modify the logo for your Awesome list or other usage.

# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6

### Styles
<style>

body.theme-dark img.toggle {
    filter: invert(100%);
}
</style>

### Scripts

<script>
ComplexDiagram(
    Terminal("basic query", {href: "#basic-query"}),
    ZeroOrMore(
        Sequence(
            Choice(0,
                Terminal("AND"),
                Terminal("OR")),
            Terminal("basic query", {href: "#basic-query"})),
        null,
        'skip')).addTo();
</script>

### Links

**Example:** [`repo:github.com/sourcegraph/sourcegraph rtr AND newRouter` ↗](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+rtr+AND+newRouter&patternType=literal)

Hello

This folder contains client code that is **branded**, i.e. it implements the visual design language we use across our web app and e.g. in the options menu of the browser extension.
Code in here can use Bootstrap and must not adapt styles of the code host (for more details, see [Styling UI in the handbook](../../doc/dev/background-information/web/styling.md)).

Any code that is code host agnostic should go into [`../shared`](../shared) instead.

#### Heading
