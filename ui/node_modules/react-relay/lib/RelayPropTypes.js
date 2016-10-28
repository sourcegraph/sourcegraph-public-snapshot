/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayPropTypes
 * 
 */

'use strict';

var PropTypes = require('react').PropTypes;

var RelayPropTypes = {
  Container: function Container(props, propName, componentName) {
    var component = props[propName];
    if (component == null) {
      return new Error(require('fbjs/lib/sprintf')('Required prop `%s` was not specified in `%s`.', propName, componentName));
    } else if (!require('./isRelayContainer')(component)) {
      return new Error(require('fbjs/lib/sprintf')('Invalid prop `%s` supplied to `%s`, expected a RelayContainer.', propName, componentName));
    }
    return null;
  },
  Environment: function Environment(props, propName, componentName) {
    var context = props[propName];
    if (!require('./isRelayEnvironment')(context)) {
      return new Error(require('fbjs/lib/sprintf')('Invalid prop/context `%s` supplied to `%s`, expected `%s` to be ' + 'an object conforming to the `RelayEnvironment` interface.', propName, componentName, context));
    }
    return null;
  },


  QueryConfig: PropTypes.shape({
    name: PropTypes.string.isRequired,
    params: PropTypes.object.isRequired,
    queries: PropTypes.object.isRequired
  })
};

module.exports = RelayPropTypes;