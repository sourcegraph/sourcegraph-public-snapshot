# element-resize-detector
Super-optimized cross-browser resize listener for elements. Up to 37x faster than related approaches (read section 5 of the [article](http://arxiv.org/pdf/1511.01223v1.pdf)).

```
npm install element-resize-detector
```

## Usage
Include the script in the browser:
```html
<script src="node_modules/element-resize-detector/dist/element-resize-detector.min.js"></script>
```
This will create a global function `elementResizeDetectorMaker`, which is the maker function that makes an element resize detector instance.

You can also `require` it like so:
```js
var elementResizeDetectorMaker = require("element-resize-detector");
```

### Create instance
```js
// With default options (will use the object-based approach).
// The object-based approach is deprecated, and will be removed in v2.
var erd = elementResizeDetectorMaker();

// With the ultra fast scroll-based approach.
// This will be the default in v2.
var erdUltraFast = elementResizeDetectorMaker({
  strategy: "scroll" //<- For ultra performance.
});
```

## API

### listenTo(element, listener)
Listens to the element for resize events and calls the listener function with the element as argument on resize events.

**Example usage:**

```js
erd.listenTo(document.getElementById("test"), function(element) {
  var width = element.offsetWidth;
  var height = element.offsetHeight;
  console.log("Size: " + width + "x" + height);
});
```

### removeListener(element, listener)
Removes the listener from the element.

### removeAllListeners(element)
Removes all listeners from the element, but does not completely remove the detector. Use this function if you may add listeners later and don't want the detector to have to initialize again.

### uninstall(element)
Completely removes the detector and all listeners.

## Caveats

1. If the element has `position: static` it will be changed to `position: relative`. Any unintentional `top/right/bottom/left/z-index` styles will therefore be applied and absolute positioned children will be positioned relative to the element.
2. A hidden element will be injected as a direct child to the element.

## Credits
This library is using the two approaches (scroll and object) as first described at [http://www.backalleycoder.com/2013/03/18/cross-browser-event-based-element-resize-detection/](http://www.backalleycoder.com/2013/03/18/cross-browser-event-based-element-resize-detection/).

The scroll based approach implementation was based on Marc J's implementation [https://github.com/marcj/css-element-queries/blob/master/src/ResizeSensor.js](https://github.com/marcj/css-element-queries/blob/master/src/ResizeSensor.js).

Please note that both approaches have been heavily reworked for better performance and robustness.

## Changelog

#### 1.1.0

* Supporting inline elements
* Event-based solution for detecting attached/rendered events so that detached/unrendered elements can be listened to without polling
* Now all changes that affects the offset size of an element are properly detected (such as padding and font-size).
* Scroll is stabilized, and is the preferred strategy to use. The object strategy will be deprecated (and is currently only used for some legacy browsers such as IE9 and Opera 12).
