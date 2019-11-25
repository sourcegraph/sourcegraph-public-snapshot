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

## Making invalid states impossible through union types

TypeScript has a very powerful type system, including support for [union types](https://www.typescriptlang.org/docs/handbook/advanced-types.html#union-types).
A union type `A | B | C` declares that the value can be either of type `A`, `B` or `C`.
Often times when thinking from the perspective of the state consumer, you end up with several different state slots that need to be represented.
For example, in a UI that fetches user data from the network, you need to show a loader when the data is loading, the contents if the result arrived, or an error if the fetch failed.
A naive approach to model this would look like this:

```ts
// BAD
{
  isLoading: boolean
  error?: Error
  user?: User
}
```

These three properties create a matrix with a total of 2<sup>3</sup> = 8 different possible states.
However, only 3 of these states are _valid_:
You want to either show the loader, the error **or** the contents, never any of them at the same time.
But by using three properties, it is possible that e.g. `isLoading` is `true`, while `user` is defined too.
The application needs to ensure that for any change of one of the properties, the other two are reset, e.g. `isLoading` is reset to `false` as soon as the contents are loaded, as well as `errorMessage`.

There are two bad effects of this:
- As the code evolves and grows more complex, these checks could easily be omitted, and the application could end up showing a loader and the contents at the same time.
- TypeScript has no way to know that `contents` is always defined when `isLoading` is not, so you'll have to either cast in multiple places or duplicate checks. The cast could then become invalid in the future and cause errors.

This can be easily avoided by expressing the mutually exclusive nature of the state slots in the type system through a union type:

```ts
// GOOD
{
  userOrError?: User | Error
}
```

Here, `undefined` represents _loading_, and a defined value means either a successful fetch **or** an error.
To find out which state is active, you can use type checking features:
- comparing to `undefined`, `null` or a defined constant with `===`
- using `typeof` if one of the states is a primitive type
- using `instanceof` if one of the states is a class
- checking a [discriminator property](https://www.typescriptlang.org/docs/handbook/advanced-types.html#discriminated-unions) like `type` or `kind` if available
- using a [custom type guard function](https://www.typescriptlang.org/docs/handbook/advanced-types.html#type-guards-and-differentiating-types) if you need to distinguish between completely different object interfaces

TypeScript is then able to narrow the type correctly in the conditional branches defined by `if`/`else`, the ternary operator, or `return`:

```ts
if (userOrError === undefined) {
  return <Loader />
}
if (userOrError instanceof Error) {
  return <div class="alert alert-danger">{upperFirst(userOrError.message)}</div>
}
return <div>Username: {userOrError.username}</div>
```

We are now relying on the type system to enforce at compile time that it is impossible to have a loader and an error shown at the same time.

**Caveat**: If you are using the `Error` type, make sure that the exception you are saving is actually an Error with our `asError()` utility.
Exceptions in TypeScript are of type `any`, i.e. if a function deep down exhibits the bad practice of throwing e.g. a string, it could mess up the type checking logic.
