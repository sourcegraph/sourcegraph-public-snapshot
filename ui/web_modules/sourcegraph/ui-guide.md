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
Containers, also called controller-views, connect stores and components. The container is the only React component that listens to store changes. On change, it fetches the new data from the stores and passes it down to child components via properties. Containers must register stores with a `stores` method and they must call the parent `componentWillMount` and `componentWillUnmount` methods.

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
