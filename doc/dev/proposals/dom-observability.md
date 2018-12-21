# Observing the DOM in the browser extension

Create a more robust way of handling DOM interactions.

## Background

I frequently see issues that are filed that just say "the browser
extension doesn't work" and then when I go to any links provided, it
works for me. I also somewhat frequently see errors come in through
Sentry that are related to elements not being found when we would
logically expect then to be there. This makes me think we should remove
any assumptions we make around the DOM. This project will fix bugs we
know exist and hopefully take care of a lot of seemingly flakey and
non-reproducible errors at the same time.

## Proposal

Create a way to observe the DOM rather than simply query the DOM. We
already have this implemented for listening for new code views being
added to the DOM. I'll abstract that implementation and use that for all
locations where we query the DOM. The API would look as follows:

```typescript
interface DOMObserver {
  /**
   * Listen to the DOM for a specific selector. It will immediately emit all
   * elements already in the DOM and will asynchronously emit new ones added.
   *
   * @param selector the selector we are wanting to match in the DOM.
   * @param timeout time in ms to wait before throwing an error if no matches are found.
   */
  observeSelector(selector: string, timeout = 500): Observable<HTMLElement>
  /**
   * Unsubscribe to either a specific selector or all selectors being listened to.
   *
   * @param selector if provided, stop listening to that specific selector. If
   * not provided, stop listening to all selectors.
   */
  unsubscribe(selector?: string): void
}
```

### Test plan

The end result will be hard to test in an automated way. It will rely on
the browser extension's e2e tests. However, I plan on creating a library
for observing the DOM and will use Karma/mocha to run unit tests for
this in actual browsers. The reason for using Karma/mocha is [JSDom
doesn't support `MutaionObserver`'s
yet.](https://github.com/jsdom/jsdom/pull/2398)

### Release plan

Nothing special here. Just release the browser extension once merged.
We'll monitor the types of errors and issues mentioned in the background
afterwards.

## Rationale

After watching issues come and observing the errors being sent to
Sentry, I believe this change is necessary. Whether its the root cause
of *all* these errors is unknown, but it will fix a lot of errors we are
seeing.

This will be it's own library rather than a part of the main Sourcegraph
repository because it is a very isolated functionality we want to provide and
should not and will not have any Sourcegraph specific logic. Once in the stable
state, it won't be updated frequently (the API is small and won't have changes).
In addition to the logical split, we want to have full freedom over the test
environment for this library and don't want to muddy up the main repo's testing
infrastructure. The main repo now uses jest and JSDom for testing which doesn't
support `MutationObserver` yet. This is why we'll likely use Karma to run tests
in real browsers.

## Implementation

- [ ] Create dom observer library
  - This uses a `MutationObserver` to listen to the DOM for changes
        and will emit when a selector matches an element in a DOM
        change.
- [ ] Write unit tests for this library.
- [ ] Replace usages of `document.querySelector` and the like in the
    browser extension. This will require a change to the
    [MountGetter](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/browser/src/libs/code_intelligence/code_intelligence.tsx#L109:8)
    type. It will either return an observable or we will just accept a
    selector and internally use our library.
