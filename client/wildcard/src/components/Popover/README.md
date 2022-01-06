## Popover/floating-panel component

This component provides a special React component (wrapper over [@floating-element package](https://floating-ui.com)) and a few low-level utilities
for building popovers/tooltips/floating-panels.

### General overview

By default `<Popover />` component requires only target element for being rendered.

```tsx
import { Button, Popover } from '@sourcegraph/wildcard'

function ComponentWithPopover() {
  const buttonRef = useRef<HTMLElement>(null)

  return (
    <section>
      <Button ref={buttonRef}>Click me</Button>

      {buttonRef.current && <Popover target={buttonRef.current}>Any popover content here</Popover>}
    </section>
  )
}
```

See _**Standard storybook example**_ of this component in `./Popover.story.tsx` for more details.

### Props overview

- `target` - reference element for popover positioning.

- `placement` - Initial popover element position. It may be changes in calculation logic finds
  a more appropriate place for the popover based on scroll and constraint elements.

- `strategy` - (absolute or fixed).

  - **With absolute** (default value) popover renders itself right
    next the target element and includes all parent elements as elements that popover should be visually fit in.
  - **With fixed** strategy popover element is rendered outside of target DOM hierarchy, and it's rendered
    in special containers on `body` element top level.

- `padding` - displaces the floating element from its reference element. See https://floating-ui.com/docs/offset for more details.

- `constraints` - List of elements that constraint position of popover element. By default, constraint is equal to
  all scrollable elements from the target and popover elements to body tag up DOM tree.

- `middlewares` - List of positioning middlewares for [@floating-element package](https://floating-ui.com). In case if you have to
  build some custom position logic completely different from this component positioning you can provide your own middlewares. See
  https://floating-ui.com/docs/middleware for more details.
