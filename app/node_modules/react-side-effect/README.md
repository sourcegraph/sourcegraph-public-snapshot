# React Side Effect

Create components whose prop changes map to a global side effect.

## Installation

```
npm install --save react-side-effect
```

Note: React Side Effect requires React 0.13+.

## Use Cases

* Setting `document.style.overflow` or background color depending on current screen;
* Firing Flux actions using declarative API depending on current screen;
* Some crazy stuff I haven't thought about.

## How's That Different from `componentDidUpdate`?

It gathers current props across *the whole tree* before passing them to side effect. For example, this allows you to create `<BodyStyle style>` component like this:

```js
// RootComponent.js
return (
  <BodyStyle style={{ backgroundColor: 'red' }}>
    {this.state.something ? <SomeComponent /> : <OtherComponent />}
  </BodyStyle>
);

// SomeComponent.js
return (
  <BodyStyle style={{ backgroundColor: this.state.color }}>
    <div>Choose color: <input valueLink={this.linkState('color')} /></div>
  </BodyStyle>
);
```

and let the effect handler merge `style` from different level of nesting with innermost winning:

```js
import { Component, Children, PropTypes } from 'react';
import withSideEffect from 'react-side-effect';

class BodyStyle extends Component {
  render() {
    return Children.only(this.props.children);
  }
}

BodyStyle.propTypes = {
  style: PropTypes.object.isRequired
};

function reducePropsToState(propsList) {
  var style = {};
  propsList.forEach(function (props) {
    Object.assign(style, props.style);
  });
  return style;
}

function handleStateChangeOnClient(style) {
  for (var key in style) {
    document.style[key] = style[key];
  }
}

export default withSideEffect(
  reducePropsToState,
  handleStateChangeOnClient
)(BodyStyle);
```

On the server, you’ll be able to call `BodyStyle.peek()` to get the current state, and `BodyStyle.rewind()` to reset for each next request. The `handleStateChangeOnClient` will only be called on the client.

## API

#### `withSideEffect: (reducePropsToState, handleStateChangeOnClient, [mapStateOnServer]) -> ReactComponent -> ReactComponent`

A [higher-order component](https://medium.com/@dan_abramov/mixins-are-dead-long-live-higher-order-components-94a0d2f9e750) that, when mounting, unmounting or receiving new props, calls `reducePropsToState` with `props` of **each mounted instance**. It is up to you to return some state aggregated from these props.

On the client, every time the returned component is (un)mounted or its props change, `reducePropsToState` will be called, and the recalculated state will be passed to `handleStateChangeOnClient` where you may use it to trigger a side effect.

On the server, `handleStateChangeOnClient` will not be called. You will still be able to call the static `rewind()` method on the returned component class to retrieve the current state after a `renderToString()` call. If you forget to call `rewind()` right after `renderToString()`, the internal instance stack will keep growing, resulting in a memory leak and incorrect information. You must call `rewind()` after every `renderToString()` call on the server.

For testing, you may use a static `peek()` method available on the returned component. It lets you get the current state without resetting the mounted instance stack. Don’t use it for anything other than testing.

## Usage

Here's how to implement [React Document Title](https://github.com/gaearon/react-document-title) (both client and server side) using React Side Effect:

```js
import React, { Children, Component, PropTypes } from 'react';
import withSideEffect from 'react-side-effect';

class DocumentTitle extends Component {
  render: function render() {
    if (this.props.children) {
      return Children.only(this.props.children);
    } else {
      return null;
    }
  }
}

DocumentTitle.propTypes = {
  title: PropTypes.string.isRequired
};

function reducePropsToState(propsList) {
  var innermostProps = propsList[propsList.length - 1];
  if (innermostProps) {
    return innermostProps.title;
  }
}

function handleStateChangeOnClient(title) {
  document.title = title || '';
}

export default withSideEffect(
  reducePropsToState,
  handleStateChangeOnClient
)(DocumentTitle);
```
