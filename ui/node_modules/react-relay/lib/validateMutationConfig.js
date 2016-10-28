/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule validateMutationConfig
 * 
 */

'use strict';

var _keys2 = _interopRequireDefault(require('babel-runtime/core-js/object/keys'));

var _freeze2 = _interopRequireDefault(require('babel-runtime/core-js/object/freeze'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var FUZZY_THRESHOLD = 3;

/* eslint-disable no-unused-vars */
var DEPRECATED = (0, _freeze2['default'])({
  assert: require('fbjs/lib/warning'),
  message: 'has deprecated property',
  type: 'DEPRECATED'
});
/* eslint-enable no-unused-vars */

var OPTIONAL = (0, _freeze2['default'])({
  // These first two properties are not needed, but including them is easier
  // than getting Flow to accept a disjoint union.
  assert: function assert() {},
  message: '',
  type: 'OPTIONAL'
});

var REQUIRED = {
  assert: require('fbjs/lib/invariant'),
  message: 'must have property',
  type: 'REQUIRED'
};

function validateMutationConfig(config, name) {
  function assertValid(properties) {
    // Check for unexpected properties.
    (0, _keys2['default'])(config).forEach(function (property) {
      if (property === 'type') {
        return;
      }

      if (!properties.hasOwnProperty(property)) {
        var _message = require('fbjs/lib/sprintf')('validateMutationConfig: Unexpected key `%s` in `%s` config ' + 'for `%s`', property, config.type, name);
        var suggestion = (0, _keys2['default'])(properties).find(function (candidate) {
          return require('./testEditDistance')(candidate, property, FUZZY_THRESHOLD);
        });
        if (suggestion) {
          require('fbjs/lib/invariant')(false, '%s; did you mean `%s`?', _message, suggestion);
        } else {
          /* eslint-disable fb-www/sprintf-like-args-uniqueness */
          require('fbjs/lib/invariant')(false, '%s.', _message);
          /* eslint-enable fb-www/sprintf-like-args-uniqueness */
        }
      }
    });

    // Check for deprecated and missing properties.
    (0, _keys2['default'])(properties).forEach(function (property) {
      var validator = properties[property];
      var isRequired = validator.type === 'REQUIRED';
      var isDeprecated = validator.type === 'DEPRECATED';
      var present = config.hasOwnProperty(property);
      if (isRequired && !present || isDeprecated && present) {
        validator.assert(false, 'validateMutationConfig: `%s` config on `%s` %s `%s`.', config.type, name, validator.message, property);
      }
    });
  }

  switch (config.type) {
    case 'FIELDS_CHANGE':
      assertValid({
        fieldIDs: REQUIRED
      });
      break;

    case 'RANGE_ADD':
      assertValid({
        connectionName: REQUIRED,
        edgeName: REQUIRED,
        parentID: OPTIONAL,
        parentName: OPTIONAL,
        rangeBehaviors: REQUIRED
      });
      break;

    case 'NODE_DELETE':
      assertValid({
        connectionName: REQUIRED,
        deletedIDFieldName: REQUIRED,
        parentID: OPTIONAL,
        parentName: REQUIRED
      });
      break;

    case 'RANGE_DELETE':
      assertValid({
        connectionName: REQUIRED,
        deletedIDFieldName: REQUIRED,
        parentID: OPTIONAL,
        parentName: REQUIRED,
        pathToConnection: REQUIRED
      });
      break;

    case 'REQUIRED_CHILDREN':
      assertValid({
        children: REQUIRED
      });
      break;

    default:
      require('fbjs/lib/invariant')(false, 'validateMutationConfig: unknown config type `%s` on `%s`', config.type, name);
  }
}

module.exports = validateMutationConfig;