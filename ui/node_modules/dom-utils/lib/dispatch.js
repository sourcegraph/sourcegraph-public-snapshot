/**
 * Dispatches an event on the passed element.
 * @param {Element} element The DOM element to dispatch the event on.
 * @param {string} eventType The type of event to dispatch.
 * @param {string} eventName (optional) A string name of the event constructor
 *     to use. Defaults to 'Event' if nothing is passed or 'CustomEvent' if
 *     a value is set on `initDict.detail`.
 * @param {Object} initDict (optional) The initialization attributes for the
 *     event. A `detail` property can be used here to pass custom data.
 * @return {boolean} The return value of `element.dispatchEvent`, which will
 *     be false if any of the event listeners called `preventDefault`.
 */
module.exports = function(element, eventType, eventName, initDict) {

  var event;
  var isCustom;

  // eventName is optional
  if (typeof eventName == 'object') {
    initDict = eventName;
    eventName = null;
  }

  // To allow for reasonable defaults, events should bubble and be cancelable
  // unless explicitly told not to.
  initDict = initDict || {};
  initDict.bubbles = 'bubbles' in initDict ? initDict.bubbles : true;
  initDict.cancelable = 'cancelable' in initDict ? initDict.cancelable : true;

  // If a detail property is passed, this is a custom event.
  if ('detail' in initDict) isCustom = true;
  eventName = isCustom ? 'CustomEvent' : eventName || 'Event';

  // Tries to create the event using constructors, if that doesn't work,
  // fallback to `document.createEvent()`.
  try {
    event = new window[eventName](eventType, initDict);
  }
  catch(err) {
    event = document.createEvent(eventName);
    var initMethod = 'init' + (isCustom ? 'Custom' : '') + 'Event';
    event[initMethod](eventType, initDict.bubbles,
                      initDict.cancelable, initDict.detail);
  }

  return element.dispatchEvent(event);
};
