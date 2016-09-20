# glamor

![build status](https://travis-ci.org/threepointone/glamor.svg)

css for component systems

`npm install glamor --save`

(or)

```html
<script src='https://unpkg.com/glamor/umd/index.min.js'></script>
```

usage 
```jsx
import { style, hover } from 'glamor'
// ...
<div 
  {...style({ color: 'red' })} 
  {...hover({ color: 'pink' })}>
  zomg
</div>
```

motivation
---

This expands on ideas from @vjeux's [2014 css-in-js talk](https://speakerdeck.com/vjeux/react-css-in-js).
We introduce an api to annotate arbitrary dom nodes with style definitions ("rules") for, um, the greater good.

features
---

- really small / fast / efficient, with a fluent api
- framework independent
- adds vendor prefixes / fallback values 
- supports all the pseudo :classes/::elements
- `@media` queries
- `@font-face` / `@keyframes`
- escape hatches for parent and child selectors 
- dev helper to simulate pseudo classes like `:hover`, etc
- server side / static rendering
- tests / coverage

(thanks to [BrowserStack](https://www.browserstack.com/) for providing the infrastructure that allows us to run our build in real browsers.)

docs 
---
- [api documentation](https://github.com/threepointone/glamor/blob/master/docs/api.md)
- [howto](https://github.com/threepointone/glamor/blob/master/docs/howto.md) - a comparison of css techniques in glamor
- [plugins](https://github.com/threepointone/glamor/blob/master/docs/plugins.md)
- [server side rendering](https://github.com/threepointone/glamor/blob/master/docs/server.md)

extras 
---

- `glamor/reset` - include a css reset
- `glamor/react` - helpers for [themes](https://github.com/threepointone/glamor/blob/master/docs/themes.md), [inline 'css' prop](https://github.com/threepointone/glamor/blob/master/docs/createElement.md), [`@vars`](https://github.com/threepointone/glamor/blob/master/docs/vars.md)
- `glamor/jsxstyle` - [react integration](https://github.com/threepointone/glamor/blob/master/docs/jsxstyle.md), à la [jsxstyle](https://github.com/petehunt/jsxstyle/)
- `glamor/aphrodite` - [shim](https://github.com/threepointone/glamor/blob/master/docs/aphrodite.md) for [aphrodite](https://github.com/Khan/aphrodite) stylesheets
- `glamor/utils` - a port of [postcss-utilities](https://github.com/ismamz/postcss-utilities)
- `glamor/ous` - a port of [the skeleton css framework](http://getskeleton.com)


characteristics
---

while glamor shares most common attributes of other inline style / css-in-js systems,
here are some key differences -

- it does **not** touch `class`/`style` attributes, neither does it **not** generate pretty classnames; instead we use `data-*` attributes and jsx attribute spread ([some implications](https://github.com/Khan/aphrodite/issues/25)). This lets you define styles 'inline', yet globally optimize as one unit.
- in dev mode, you can simulate pseudo classes on elements by using the `simulate()` helper (or adding a `[data-simulate-<pseudo>]` attribute manually). very useful, especially when combined when hot-loading and/or editing directly in your browser devtools.
- in production, we switch to a **much** faster method of appending styles to the document, able to add 10s of 1000s of rules in milliseconds. the caveat with this approach is that those styles will [not be editable in chrome/firefox devtools](https://bugs.chromium.org/p/chromium/issues/detail?id=387952) (which is fine, for prod?). advanced users may use `speedy(true/false)` to toggle this setting manually. 


todo
---

- [planned enhancements](https://github.com/threepointone/glamor/issues?q=is%3Aopen+is%3Aissue+label%3Aenhancement)
- notes on composition
- ie8 compat

profit, profit
---

I get it
