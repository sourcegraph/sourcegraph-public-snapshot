# react-draggable [![Build Status](https://travis-ci.org/mzabriskie/react-draggable.svg?branch=master)](https://travis-ci.org/mzabriskie/react-draggable)

A simple component for making elements draggable.

[View the Changelog](CHANGELOG.md)

## Demo

[View Demo](http://mzabriskie.github.io/react-draggable/example/)


## Installing

```bash
$ npm install react-draggable
```

If you aren't using browserify/webpack, a
[UMD version of react-draggable](dist/react-draggable.js) is available. It is updated per-release only.

If you want a UMD version of the latest `master` revision, you can generate it yourself from master by cloning this
repository and running `$ make`. This will create umd dist files in the `dist/` folder.

## Details

A `<Draggable>` element wraps an existing element and extends it with new event handlers and styles.
It does not create a wrapper element in the DOM.

Draggable items are moved using CSS Transforms. This allows items to be dragged regardless of their current
positioning (relative, absolute, or static). Elements can also be moved between drags without incident.

If the item you are dragging already has a CSS Transform applied, it will be overwritten by `<Draggable>`. Use
an intermediate wrapper (`<Draggable><span>...</span></Draggable>`) in this case.

## API
The `<Draggable/>` component transparently adds draggable to whatever element is supplied as `this.props.children`.
**Note**: Only a single element is allowed or an Error will be thrown.

Props:

**`axis`**: determines which axis the draggable can move. Accepted values:
- `both` allows movement horizontally and vertically (default).
- `x` limits movement to horizontal axis.
- `y` limits movement to vertical axis.

**`handle`**: specifies a selector to be used as the handle that initiates drag.

**`cancel`**: specifies a selector to be used to prevent drag initialization.

**`grid`**: specifies the x and y that dragging should snap to.

**`bounds`**: specifies movement boundaries. Accepted values:
- `parent` restricts movement within the node's offsetParent (nearest node with position relative or absolute), or
- An object with `left, top, right, and bottom` properties. These indicate how far in each direction the draggable can be moved. See [example/index.html](https://github.com/mzabriskie/react-draggable/blob/master/example/index.html) for more on this.

**`start`**: specifies the `x` and `y` that the dragged item should start at. This is generally not necessary to use (you can use absolute or relative positioning of the child directly), but can be helpful for uniformity in your callbacks and with css transforms.

**`moveOnStartChange`**: if true (it defaults false), will move the element if there is a change in `start`. We set this by default to `false` because it can cause unwanted effects if you are not aware of it.

**`zIndex`**: specifies the zIndex to use while dragging.

**`onStart`**: called when dragging starts.

**`onDrag`**: called while dragging.

**`onStop`**: called when dragging stops.


## Example usage

```js
/** @jsx React.DOM */
var React = require('react'),
	Draggable = require('react-draggable');

var App = React.createClass({
	handleStart: function (event, ui) {
		console.log('Event: ', event);
		console.log('Position: ', ui.position);
	},

	handleDrag: function (event, ui) {
		console.log('Event: ', event);
        console.log('Position: ', ui.position);
	},

	handleStop: function (event, ui) {
		console.log('Event: ', event);
        console.log('Position: ', ui.position);
	},

	render: function () {
		return (
			<Draggable
				axis="x"
				handle=".handle"
				start={{x: 0, y: 0}}
				moveOnStartChange={false}
				grid={[25, 25]}
				zIndex={100}
				onStart={this.handleStart}
				onDrag={this.handleDrag}
				onStop={this.handleStop}>
				<div>
					<div className="handle">Drag from here</div>
					<div>This readme is really dragging on...</div>
				</div>
			</Draggable>
		);
	}
});

React.renderComponent(<App/>, document.body);
```

## State Problems?

`<Draggable>` is a stateful component. This means that it is storing its current drag offsets in its internal state.
This can cause problems with certain integrations. For example, if you change the position of the element manually,
`<Draggable>` can get into trouble as it assumes a translation in the DOM. If you see an element jump around the page
when you click it, this is affecting you.

This is an unfortunate side-effect of dragging, which is inherently stateful.

If you move the element manually, you have two options:

1. Feed the `<Draggable>` an `x` and `y` parameter in the `start` param, and change it as you go while setting
`moveOnStartChange` to `true`, or,
2. When moving the `<Draggable>`, ref the element and
[call `resetState()`](https://github.com/STRML/react-resizable/blob/master/lib/Resizable.jsx#L48).


## Contributing

- Fork the project
- Run the project in development mode: `$ make dev`
- Make changes.
- Add appropriate tests
- `$ make test`
- If tests don't pass, make them pass.
- Update README with appropriate docs.
- Commit and PR

## Release checklist

- Update CHANGELOG
- `make release-patch`, `make release-minor`, or `make-release-major`
- `make publish`

## License

MIT
