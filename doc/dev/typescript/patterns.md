# TypeScript programming patterns

This is a non-exhaustive lists of programming patterns used across our TypeScript codebases.
Since patterns are used everywhere, they are usually not documented with inline code comments.
This document serves as the canonical place to document them.

## `Subscription` bag

Classes or React components often register external subscriptions, usually with RxJS.
These Subscriptions must be cleaned up when the instance of the class is no longer in use,
but won't be automatically through garbage collection, because the external emitter holds a (transitive) reference to the class instance, not the other way around.

If the Subscriptions are not cleaned up, the garbage collector can not collect the class instance,
and the subscribe callback may still run code operating on invalid state.
In the case of React components, it may call `setState()` after the component was unmounted,
which causes React to throw an Error.

If you forgot to handle a Subscription returned by Rx `observable.subscribe()`, the TSLint rule `rxjs-no-ignored-subscription` will warn you.

The easiest way to solve this is to save all Subscriptions in a Subscription bag, and when the instance is disposed, unsubscribe that bag.
[RxJS Subscriptions](https://rxjs-dev.firebaseapp.com/guide/subscription) have the nice property of being _composable_, meaning they can act as a Subscription bag.
You can add more Subscriptions to it with the `add()` method, and unsubscribe all at once by calling `unsubscribe()`.
It handles all the edge cases like adding a Subscription after it was already unsubscribed.

Subscriptions don't just accept other Subscriptions, they accept arbitrary cleanup functions to be run on unsubscription, so this pattern can also be used to clean up non-RxJS resources or listeners.

Example of what this pattern typically looks like in a React component:

```ts
class MyComponent extends React.Component {
  // Subscription used as a Subscription bag
  private subscriptions = new Subscription()

  public componentDidMount(): void {
    // Register the Subscription in the Subscription bag
    this.subscriptions.add(
      someObservable.subscribe(value => {
        // This is guaranteed to never run after this.subscriptions.unsubscribe() was called
        this.setState({ value })
      })
    )
  }

  // React lifecycle method that gets run on disposal
  public componentWillUnmount(): void {
    // Unsubscribes all Subscriptions that were made
    this.subscriptions.unsubscribe()
  }
}
```
