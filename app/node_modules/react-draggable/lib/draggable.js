'use strict';

var React = require('react');
var emptyFunction = function(){};
var assign = require('object-assign');
var classNames = require('classnames');

//
// Helpers. See Element definition below this section.
//

function createUIEvent(draggable) {
  // State changes are often (but not always!) async. We want the latest value.
  var state = draggable._pendingState || draggable.state;
  return {
    node: draggable.getDOMNode(),
    position: {
      top: state.clientY,
      left: state.clientX
    }
  };
}

function canDragY(draggable) {
  return draggable.props.axis === 'both' ||
      draggable.props.axis === 'y';
}

function canDragX(draggable) {
  return draggable.props.axis === 'both' ||
      draggable.props.axis === 'x';
}

function isFunction(func) {
  return typeof func === 'function' || Object.prototype.toString.call(func) === '[object Function]';
}

// @credits https://gist.github.com/rogozhnikoff/a43cfed27c41e4e68cdc
function findInArray(array, callback) {
  for (var i = 0, length = array.length; i < length; i++) {
    if (callback.apply(callback, [array[i], i, array])) return array[i];
  }
}

function matchesSelector(el, selector) {
  var method = findInArray([
    'matches',
    'webkitMatchesSelector',
    'mozMatchesSelector',
    'msMatchesSelector',
    'oMatchesSelector'
  ], function(method){
    return isFunction(el[method]);
  });

  return el[method].call(el, selector);
}

/**
 * simple abstraction for dragging events names
 * */
var eventsFor = {
  touch: {
    start: 'touchstart',
    move: 'touchmove',
    end: 'touchend'
  },
  mouse: {
    start: 'mousedown',
    move: 'mousemove',
    end: 'mouseup'
  }
};

// Default to mouse events
var dragEventFor = eventsFor['mouse'];

/**
 * get {clientX, clientY} positions of control
 * */
function getControlPosition(e) {
  var position = (e.touches && e.touches[0]) || e;
  return {
    clientX: position.clientX,
    clientY: position.clientY
  };
}

function addEvent(el, event, handler) {
  if (!el) { return; }
  if (el.attachEvent) {
    el.attachEvent('on' + event, handler);
  } else if (el.addEventListener) {
    el.addEventListener(event, handler, true);
  } else {
    el['on' + event] = handler;
  }
}

function removeEvent(el, event, handler) {
  if (!el) { return; }
  if (el.detachEvent) {
    el.detachEvent('on' + event, handler);
  } else if (el.removeEventListener) {
    el.removeEventListener(event, handler, true);
  } else {
    el['on' + event] = null;
  }
}

function outerHeight(node) {
  // This is deliberately excluding margin for our calculations, since we are using
  // offsetTop which is including margin. See getBoundPosition
  var height = node.clientHeight;
  var computedStyle = window.getComputedStyle(node);
  height += int(computedStyle.borderTopWidth);
  height += int(computedStyle.borderBottomWidth);
  return height;
}

function outerWidth(node) {
  // This is deliberately excluding margin for our calculations, since we are using
  // offsetLeft which is including margin. See getBoundPosition
  var width = node.clientWidth;
  var computedStyle = window.getComputedStyle(node);
  width += int(computedStyle.borderLeftWidth);
  width += int(computedStyle.borderRightWidth);
  return width;
}
function innerHeight(node) {
  var height = node.clientHeight;
  var computedStyle = window.getComputedStyle(node);
  height -= int(computedStyle.paddingTop);
  height -= int(computedStyle.paddingBottom);
  return height;
}

function innerWidth(node) {
  var width = node.clientWidth;
  var computedStyle = window.getComputedStyle(node);
  width -= int(computedStyle.paddingLeft);
  width -= int(computedStyle.paddingRight);
  return width;
}

function isNum(num) {
  return typeof num === 'number' && !isNaN(num);
}

function int(a) {
  return parseInt(a, 10);
}

function getBoundPosition(draggable, clientX, clientY) {
  var bounds = JSON.parse(JSON.stringify(draggable.props.bounds));
  var node = draggable.getDOMNode();
  var parent = node.parentNode;

  if (bounds === 'parent') {
    var nodeStyle = window.getComputedStyle(node);
    var parentStyle = window.getComputedStyle(parent);
    // Compute bounds. This is a pain with padding and offsets but this gets it exactly right.
    bounds = {
      left: -node.offsetLeft + int(parentStyle.paddingLeft) +
            int(nodeStyle.borderLeftWidth) + int(nodeStyle.marginLeft),
      top: -node.offsetTop + int(parentStyle.paddingTop) +
            int(nodeStyle.borderTopWidth) + int(nodeStyle.marginTop),
      right: innerWidth(parent) - outerWidth(node) - node.offsetLeft,
      bottom: innerHeight(parent) - outerHeight(node) - node.offsetTop
    };
  }

  // Keep x and y below right and bottom limits...
  if (isNum(bounds.right)) clientX = Math.min(clientX, bounds.right);
  if (isNum(bounds.bottom)) clientY = Math.min(clientY, bounds.bottom);

  // But above left and top limits.
  if (isNum(bounds.left)) clientX = Math.max(clientX, bounds.left);
  if (isNum(bounds.top)) clientY = Math.max(clientY, bounds.top);

  return [clientX, clientY];
}

function snapToGrid(grid, pendingX, pendingY) {
  var x = Math.round(pendingX / grid[0]) * grid[0];
  var y = Math.round(pendingY / grid[1]) * grid[1];
  return [x, y];
}

// Useful for preventing blue highlights all over everything when dragging.
var userSelectStyle = ';user-select: none;-webkit-user-select:none;-moz-user-select:none;' +
  '-o-user-select:none;-ms-user-select:none;';

function addUserSelectStyles(draggable) {
  if (!draggable.props.enableUserSelectHack) return;
  var style = document.body.getAttribute('style') || '';
  document.body.setAttribute('style', style + userSelectStyle);
}

function removeUserSelectStyles(draggable) {
  if (!draggable.props.enableUserSelectHack) return;
  var style = document.body.getAttribute('style') || '';
  document.body.setAttribute('style', style.replace(userSelectStyle, ''));
}

function createCSSTransform(style) {
  // Replace unitless items with px
  var x = style.x + 'px';
  var y = style.y + 'px';
  return {
    transform: 'translate(' + x + ',' + y + ')',
    WebkitTransform: 'translate(' + x + ',' + y + ')',
    OTransform: 'translate(' + x + ',' + y + ')',
    msTransform: 'translate(' + x + ',' + y + ')',
    MozTransform: 'translate(' + x + ',' + y + ')'
  };
}


//
// End Helpers.
//

//
// Define <Draggable>
//

module.exports = React.createClass({
  displayName: 'Draggable',

  propTypes: {
    /**
     * `axis` determines which axis the draggable can move.
     *
     * 'both' allows movement horizontally and vertically.
     * 'x' limits movement to horizontal axis.
     * 'y' limits movement to vertical axis.
     *
     * Defaults to 'both'.
     */
    axis: React.PropTypes.oneOf(['both', 'x', 'y']),

    /**
     * `bounds` determines the range of movement available to the element.
     * Available values are:
     *
     * 'parent' restricts movement within the Draggable's parent node.
     *
     * Alternatively, pass an object with the following properties, all of which are optional:
     *
     * {left: LEFT_BOUND, right: RIGHT_BOUND, bottom: BOTTOM_BOUND, top: TOP_BOUND}
     *
     * All values are in px.
     *
     * Example:
     *
     * ```jsx
     *   var App = React.createClass({
     *       render: function () {
     *         return (
     *            <Draggable bounds={{right: 300, bottom: 300}}>
     *              <div>Content</div>
     *           </Draggable>
     *         );
     *       }
     *   });
     * ```
     */
    bounds: React.PropTypes.oneOfType([
      React.PropTypes.shape({
        left: React.PropTypes.Number,
        right: React.PropTypes.Number,
        top: React.PropTypes.Number,
        bottom: React.PropTypes.Number
      }),
      React.PropTypes.oneOf(['parent', false])
    ]),

    /**
     * By default, we add 'user-select:none' attributes to the document body
     * to prevent ugly text selection during drag. If this is causing problems
     * for your app, set this to `false`.
     */
    enableUserSelectHack: React.PropTypes.bool,

    /**
     * `handle` specifies a selector to be used as the handle that initiates drag.
     *
     * Example:
     *
     * ```jsx
     *   var App = React.createClass({
     *       render: function () {
     *         return (
     *            <Draggable handle=".handle">
     *              <div>
     *                  <div className="handle">Click me to drag</div>
     *                  <div>This is some other content</div>
     *              </div>
     *           </Draggable>
     *         );
     *       }
     *   });
     * ```
     */
    handle: React.PropTypes.string,

    /**
     * `cancel` specifies a selector to be used to prevent drag initialization.
     *
     * Example:
     *
     * ```jsx
     *   var App = React.createClass({
     *       render: function () {
     *           return(
     *               <Draggable cancel=".cancel">
     *                   <div>
     *                     <div className="cancel">You can't drag from here</div>
     *            <div>Dragging here works fine</div>
     *                   </div>
     *               </Draggable>
     *           );
     *       }
     *   });
     * ```
     */
    cancel: React.PropTypes.string,

    /**
     * `grid` specifies the x and y that dragging should snap to.
     *
     * Example:
     *
     * ```jsx
     *   var App = React.createClass({
     *       render: function () {
     *           return (
     *               <Draggable grid={[25, 25]}>
     *                   <div>I snap to a 25 x 25 grid</div>
     *               </Draggable>
     *           );
     *       }
     *   });
     * ```
     */
    grid: React.PropTypes.arrayOf(React.PropTypes.number),

    /**
     * `start` specifies the x and y that the dragged item should start at
     *
     * Example:
     *
     * ```jsx
     *      var App = React.createClass({
     *          render: function () {
     *              return (
     *                  <Draggable start={{x: 25, y: 25}}>
     *                      <div>I start with transformX: 25px and transformY: 25px;</div>
     *                  </Draggable>
     *              );
     *          }
     *      });
     * ```
     */
    start: React.PropTypes.shape({
      x: React.PropTypes.number,
      y: React.PropTypes.number
    }),

    /**
     * `moveOnStartChange`, if true (default false) will move the element if the `start`
     * property changes.
     */
    moveOnStartChange: React.PropTypes.bool,


    /**
     * `zIndex` specifies the zIndex to use while dragging.
     *
     * Example:
     *
     * ```jsx
     *   var App = React.createClass({
     *       render: function () {
     *           return (
     *               <Draggable zIndex={100}>
     *                   <div>I have a zIndex</div>
     *               </Draggable>
     *           );
     *       }
     *   });
     * ```
     */
    zIndex: React.PropTypes.number,

    /**
     * Called when dragging starts.
     * If this function returns the boolean false, dragging will be canceled.
     *
     * Example:
     *
     * ```js
     *  function (event, ui) {}
     * ```
     *
     * `event` is the Event that was triggered.
     * `ui` is an object:
     *
     * ```js
     *  {
     *    position: {top: 0, left: 0}
     *  }
     * ```
     */
    onStart: React.PropTypes.func,

    /**
     * Called while dragging.
     * If this function returns the boolean false, dragging will be canceled.
     *
     * Example:
     *
     * ```js
     *  function (event, ui) {}
     * ```
     *
     * `event` is the Event that was triggered.
     * `ui` is an object:
     *
     * ```js
     *  {
     *    position: {top: 0, left: 0}
     *  }
     * ```
     */
    onDrag: React.PropTypes.func,

    /**
     * Called when dragging stops.
     *
     * Example:
     *
     * ```js
     *  function (event, ui) {}
     * ```
     *
     * `event` is the Event that was triggered.
     * `ui` is an object:
     *
     * ```js
     *  {
     *    position: {top: 0, left: 0}
     *  }
     * ```
     */
    onStop: React.PropTypes.func,

    /**
     * A workaround option which can be passed if onMouseDown needs to be accessed,
     * since it'll always be blocked (due to that there's internal use of onMouseDown)
     */
    onMouseDown: React.PropTypes.func,
  },

  componentWillReceiveProps: function(newProps) {
    // React to changes in the 'start' param.
    if (newProps.moveOnStartChange && newProps.start) {
      this.setState(this.getInitialState(newProps));
    }
  },

  componentWillUnmount: function() {
    // Remove any leftover event handlers
    removeEvent(document, dragEventFor['move'], this.handleDrag);
    removeEvent(document, dragEventFor['end'], this.handleDragEnd);
    removeUserSelectStyles(this);
  },

  getDefaultProps: function () {
    return {
      axis: 'both',
      bounds: false,
      handle: null,
      cancel: null,
      grid: null,
      moveOnStartChange: false,
      start: {x: 0, y: 0},
      zIndex: NaN,
      enableUserSelectHack: true,
      onStart: emptyFunction,
      onDrag: emptyFunction,
      onStop: emptyFunction,
      onMouseDown: emptyFunction
    };
  },

  getInitialState: function (props) {
    // Handle call from CWRP
    props = props || this.props;
    return {
      // Whether or not we are currently dragging.
      dragging: false,

      // Offset between start top/left and mouse top/left while dragging.
      offsetX: 0, offsetY: 0,

      // Current transform x and y.
      clientX: props.start.x, clientY: props.start.y
    };
  },

  handleDragStart: function (e) {
    // Make it possible to attach event handlers on top of this one
    this.props.onMouseDown(e);

    // Short circuit if handle or cancel prop was provided and selector doesn't match
    if ((this.props.handle && !matchesSelector(e.target, this.props.handle)) ||
      (this.props.cancel && matchesSelector(e.target, this.props.cancel))) {
      return;
    }

    // Call event handler. If it returns explicit false, cancel.
    var shouldStart = this.props.onStart(e, createUIEvent(this));
    if (shouldStart === false) return;

    var dragPoint = getControlPosition(e);

    // Add a style to the body to disable user-select. This prevents text from
    // being selected all over the page.
    addUserSelectStyles(this);

    // Initiate dragging. Set the current x and y as offsets
    // so we know how much we've moved during the drag. This allows us
    // to drag elements around even if they have been moved, without issue.
    this.setState({
      dragging: true,
      offsetX: dragPoint.clientX - this.state.clientX,
      offsetY: dragPoint.clientY - this.state.clientY
    });


    // Add event handlers
    addEvent(document, dragEventFor['move'], this.handleDrag);
    addEvent(document, dragEventFor['end'], this.handleDragEnd);
  },

  handleDragEnd: function (e) {
    // Short circuit if not currently dragging
    if (!this.state.dragging) {
      return;
    }

    removeUserSelectStyles(this);

    // Turn off dragging
    this.setState({
      dragging: false
    });

    // Call event handler
    this.props.onStop(e, createUIEvent(this));

    // Remove event handlers
    removeEvent(document, dragEventFor['move'], this.handleDrag);
    removeEvent(document, dragEventFor['end'], this.handleDragEnd);
  },

  handleDrag: function (e) {
    var dragPoint = getControlPosition(e);

    // Calculate X and Y
    var clientX = dragPoint.clientX - this.state.offsetX;
    var clientY = dragPoint.clientY - this.state.offsetY;

    // Snap to grid if prop has been provided
    if (Array.isArray(this.props.grid)) {
      var coords = snapToGrid(this.props.grid, clientX, clientY);
      clientX = coords[0], clientY = coords[1];
    }

    if (this.props.bounds) {
      var pos = getBoundPosition(this, clientX, clientY);
      clientX = pos[0], clientY = pos[1];
    }

    // Call event handler. If it returns explicit false, cancel.
    var shouldUpdate = this.props.onDrag(e, createUIEvent(this));
    if (shouldUpdate === false) return this.handleDragEnd();

    // Update transform
    this.setState({
      clientX: clientX,
      clientY: clientY
    });
  },

  onMouseDown: function(ev) {
    // Prevent 'ghost click' which happens 300ms after touchstart if the event isn't cancelled.
    // We don't cancel the event on touchstart because of #37; we might want to make a scrollable item draggable.
    // More on ghost clicks: http://ariatemplates.com/blog/2014/05/ghost-clicks-in-mobile-browsers/
    if (dragEventFor == eventsFor['touch']) {
      return ev.preventDefault();
    }

    return this.handleDragStart.apply(this, arguments);
  },

  onTouchStart: function(ev) {
    // We're on a touch device now, so change the event handlers
    dragEventFor = eventsFor['touch'];

    return this.handleDragStart.apply(this, arguments);
  },

  // Intended for use by a parent component. Resets internal state on this component. Useful for
  // <Resizable> and other components in case this element is manually resized and start/moveOnStartChange
  // don't work for you.
  resetState: function() {
    this.setState({
      offsetX: 0, offsetY: 0, clientX: 0, clientY: 0
    });
  },

  render: function () {
    // Create style object. We extend from existing styles so we don't
    // remove anything already set (like background, color, etc).
    var childStyle = this.props.children.props.style || {};

    // Add a CSS transform to move the element around. This allows us to move the element around
    // without worrying about whether or not it is relatively or absolutely positioned.
    // If the item you are dragging already has a transform set, wrap it in a <span> so <Draggable>
    // has a clean slate.
    var transform = createCSSTransform({
      // Set left if horizontal drag is enabled
      x: canDragX(this) ?
        this.state.clientX :
        0,

      // Set top if vertical drag is enabled
      y: canDragY(this) ?
        this.state.clientY :
        0
    });

    // Workaround IE pointer events; see #51
    // https://github.com/mzabriskie/react-draggable/issues/51#issuecomment-103488278
    var touchHacks = {
      touchAction: 'none'
    };

    var style = assign({}, childStyle, transform, touchHacks);

    // Set zIndex if currently dragging and prop has been provided
    if (this.state.dragging && !isNaN(this.props.zIndex)) {
      style.zIndex = this.props.zIndex;
    }

    var className = classNames((this.props.children.props.className || ''), 'react-draggable', {
      'react-draggable-dragging': this.state.dragging,
      'react-draggable-dragged': this.state.dragged
    });

    // Reuse the child provided
    // This makes it flexible to use whatever element is wanted (div, ul, etc)
    return React.cloneElement(React.Children.only(this.props.children), {
      style: style,
      className: className,

      onMouseDown: this.onMouseDown,
      onTouchStart: this.onTouchStart,
      onMouseUp: this.handleDragEnd,
      onTouchEnd: this.handleDragEnd
    });
  }
});
