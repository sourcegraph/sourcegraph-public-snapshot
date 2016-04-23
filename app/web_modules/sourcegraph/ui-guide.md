UI Guide
========

React and Flux
--------------
Our UI is based on the React library and the Flux architecture, see
* [React Tutorial](http://facebook.github.io/react/docs/tutorial.html)
* [React Component Specs and Lifecycle](http://facebook.github.io/react/docs/component-specs.html)
* [Flux Overview](https://facebook.github.io/flux/docs/overview.html)
* [What is Flux?](http://fluxxor.com/what-is-flux.html)

### Actions
We implement actions via ES6 classes. The class identifies the action type, not a string. They are just data containers, so they only have a constructor. We don't use action creators. Actions can be created and passed to the global dispatcher directly. Actions get created by components and backends. They get consumed by backends and stores.

### Backends
In addition to dispatcher, stores, components and actions, our architecture has another building block: Backends. They register with the dispatcher, just like stores, but they don't contain state.

All communication with the server API has to be done in a backend. Components don't talk to the backend directly, but instead create `Want*` actions that the backend responds to. When data was fetched from the server, the backend creates a `*Fetched` action to pass data to a store which persists it in its state and makes it available to the components. A `Want*` action may even be created if the data is already available. The backend then decides if it will not do anything and just keep the data which is currently in the store or if it will do another request to the server to update the data.

### Containers
Containers, also called controller-views, connect stores and components. The container is the only React component that listens to store changes. On change, it fetches the new data from the stores and passes it down to child components via properties.

### Summary
When displaying a component with data from the server, the following happens:
1. React mounts the component
2. The component renders with the currently available data, initially none
3. The component creates a `Want*` action
4. The backend responds to that action by doing a request to the server
5. When the data arrives from the server, the backend creates a `*Fetched` action
6. The store takes the data from that action and merges it into its state
7. The container gets notified that the store changed, fetches the new data from the store and passes it down to child components
8. The component renders again, this time the data is available

Code style rules
----------------
* Code from the `web_modules/sourcegraph` directory can not have dependencies to outside of this directory. To use old code, copy or move it into the `web_modules/sourcegraph` directory after making sure that it fits well.
* Prefix private methods with an underscore. Those methods should only be called on `this`, e.g. `this._foo()`.
* Use `null` instead of `undefined`. The goal is to keep `undefined` as an indicator of broken code.
* Do not use jQuery, instead use ES6 functions (we have polyfills) and npm packages that have small scope.
* Do not use Backbone.

Testing
-------
* Add tests for all actions.
* The `autotest` tool is very helpful for testing React components.
* Don't use `sinon` or other methods to override behavior of code "from the outside". Instead use proper encapsulation and some helper API for testing.
* Tests do not have to use correct types for mock data if the tested code does not care about the type, e.g. if it just passes the data through.

Development notes
-----------------

Run `enableActionLog()` in the browser JS console to log action dispatches. Run `disableActionLog()` to turn logging off.

The `Component` base class
--------------------------

    Data Flow Diagram
    =================

          Parent Component      Events
                |                 |
    +-----------|-----------------|-----------------------------+
    |           |                 |                             |
    |           v                 v                  Component  |
    |                                                           |
    |       Properties    Call to setState()                    |
    |           |                 |                             |
    |           |                 |                             |
    |           |                 |                             |
    |           |                 v                             |
    |           |                                               |
    |           +--------> reconcileState()                     |
    |                             |                             |
    |                             | ------> onStateTransition() |
    |                             |                 |           |
    |                             v                 |           |
    |                                               |           |
    |                           State               |           |
    |                             |                 |           |
    |                             |                 |           |
    |                             |                 |           |
    |                             v                 |           |
    |                                               |           |
    |                          render()             |           |
    |                             |                 |           |
    +-----------------------------|-----------------|-----------+
                                  |                 |
                                  v                 v

                            Render Output      Want* Actions

* Implement `reconcileState(state, props)` to merge new `props` into `state` and update `state` according to UI logic
* Implement `onStateTransition(prevState, nextState)` to create `Want*` actions.
* Do not access `this.props` at all.
* The component gets rendered if the state changes (shallow comparison).


## Setting statuses and loading state

Each "page" in the application must report whether any errors occurred
while loading, and when loading is complete. This is necessary for
server-side rendering, which requires that we wait until a page is
done loading and then send back an HTTP status code for the page.

You need to do 2 things to properly track the status and loading
state: record errors, and record async operations. The following
sections describe how to do each one of these things. See the
`sourcegraph/app/status` module for more information.

### Recording errors

In any React component that fetches data and that you wish to
participate in setting the success/failure state, call
`this.context.status.error(err: Error)` with any errors. As described
in the `error` method's documentation, a subsequent null error will
not overwrite a non-null error, so you needn't worry about clobbering
errors. For example:

``` javascript
class MyComponent extends React.Component {
	contextTypes = {
		status: React.PropTypes.object,
	};

	onStateTransition(prevState, nextState) {
		if (nextState.foo && prevState.foo !== nextState.foo) {
			// NOTE: Assumes that the backend stores an error in the Error field.
			this.context.status.error(nextState.foo.Error);
		}
	}
}
```

### Recording promises

Finally, every async operation should be wrapped in `trackPromise`, so
the server knows when all of them are completed (and the page can be
rendered):

``` javascript
import {trackPromise} from "sourcegraph/app/status";

const MyBackend = {
	// ...

	__onDispatch(action) {
		switch (action.constructor) {

		case MyActions.WantFoo:
			{
				let foo = FooStore.foos.get(action.foo);
				if (foo === null) {
					trackPromise(FooBackend.fetch(`/.api/foo/${action.foo}`)
						.then(...));
		// ...
	}
}
```

