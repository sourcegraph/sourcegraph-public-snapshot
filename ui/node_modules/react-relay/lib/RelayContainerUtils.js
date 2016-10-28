/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayContainerUtils
 * 
 */

'use strict';

/**
 * @internal
 *
 * Helper for checking if this is a React Component
 * created with React.Component or React.createClass().
 */

function isReactComponent(component) {
  return !!(component && component.prototype && component.prototype.isReactComponent);
}

function getReactComponent(Component) {
  if (isReactComponent(Component)) {
    return Component;
  } else {
    return null;
  }
}

function getComponentName(Component) {
  var name = void 0;
  var ComponentClass = getReactComponent(Component);
  if (ComponentClass) {
    name = ComponentClass.displayName || ComponentClass.name;
  } else if (typeof Component === 'function') {
    // This is a stateless functional component.
    name = Component.displayName || Component.name || 'StatelessComponent';
  } else {
    name = 'ReactElement';
  }
  return name;
}

module.exports = {
  getComponentName: getComponentName,
  getReactComponent: getReactComponent
};