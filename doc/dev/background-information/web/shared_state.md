# How to use shared state in React

[zustand](https://zustand.surge.sh/) is a state management library for React and
works without component state or React context.

NOTE: We are still evaluating [zustand](https://zustand.surge.sh/) and do not
recommend for general use yet.

## Why/when should I use zustand?

If you have data that needs to be accessible by multiple, but disconnected,
components. We currently store such data in the state of a common ancestor
component, but updating the data will cause intermediate in question to
re-render. Additionally the state of the ancestor component and the props of all
intermediate components get "polluted".

Zustand solves this by providing *stores* that lives outside the render tree.
Interested components subscript to stores via hooks. All that is done with very
little boilerplate code.

## Create a store

Stores contain data and "actions" (functions to update the data in the store).

```js
// useVisitorCounter.ts
import create from 'zustand';

interface VisitorStore {
  counter: number;
  increase: () => void;
  reset: () => void;
  log: () => void;
}

export const useVisitorCounter = create<VisitorStore>((set, get) => ({
  counter: 0,
  increase: () => set(state => ({counter: state.counter + 1})),
  reset: () => set({counter: 0}),
  log: () => {
    // Do whatever with the current state
    console.log(get().counter);
  }
});
```

`create` passes functions to the callback for writing to and reading from the
store:

  - `set` accepts a partial state object (or a function returning a partial
      state object) and will merge that value into the current state. 
  - `get` returns the current state object. This is useful for actions that
      don't mutate the state.

## Use the store

`useVisitorCounter` is a hook which returns the full state object. Whenever the
state changes, the component will re-render. The preferred way is to use it with
a selector, to only extract the data from the store that is needed. This way the
component will only re-render if *that* data changes.

```jsx
import {useVisitorCounter} from './useVisitorCounter';

export function MyComponent() {
  const counter = useVisitorCounter(state => state.counter);
  const increase = useVisitorCounter(state => state.increase);

  return (
    <>
      <span>{counter}</span>
      <button onClick={increase}>Increase</button>
    </>
  );
}
```

You can extract multiple values at once via the selector, in which you most
likely want to use a different comparison function. The [official README][1]
has a lot of examples.

## Use in tests

Note that the state of zustand stores doesn't reset between test runs. So it is
good to initialize the state to the expected value before each test.

The state of a store can be changed outside of React with the `setState` method
of the store:

```js
// someTest.ts
import {useVisitorCounter} from './useVisitorCounter';

// ...
useVisitorCounter.setState({counter: 42});
```

---

This only scratches the surface. As already mentioned, the [official README][1]
describes many more use cases and situations.

[1]: https://github.com/pmndrs/zustand/blob/main/readme.md
