/**
 * Copyright 2016-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule ReactDebugInstanceMap
 */

'use strict';

var warning = require('fbjs/lib/warning');

function checkValidInstance(internalInstance) {
  if (!internalInstance) {
    process.env.NODE_ENV !== 'production' ? warning(false, 'There is an internal error in the React developer tools integration. ' + 'Instead of an internal instance, received %s. ' + 'Please report this as a bug in React.', internalInstance) : void 0;
    return false;
  }
  var isValid = typeof internalInstance.mountComponent === 'function';
  process.env.NODE_ENV !== 'production' ? warning(isValid, 'There is an internal error in the React developer tools integration. ' + 'Instead of an internal instance, received an object with the following ' + 'keys: %s. Please report this as a bug in React.', Object.keys(internalInstance).join(', ')) : void 0;
  return isValid;
}

var idCounter = 1;
var instancesByIDs = {};
var instancesToIDs;

function getIDForInstance(internalInstance) {
  if (!instancesToIDs) {
    instancesToIDs = new WeakMap();
  }
  if (instancesToIDs.has(internalInstance)) {
    return instancesToIDs.get(internalInstance);
  } else {
    var instanceID = (idCounter++).toString();
    instancesToIDs.set(internalInstance, instanceID);
    return instanceID;
  }
}

function getInstanceByID(instanceID) {
  return instancesByIDs[instanceID] || null;
}

function isRegisteredInstance(internalInstance) {
  var instanceID = getIDForInstance(internalInstance);
  if (instanceID) {
    return instancesByIDs.hasOwnProperty(instanceID);
  } else {
    return false;
  }
}

function registerInstance(internalInstance) {
  var instanceID = getIDForInstance(internalInstance);
  if (instanceID) {
    instancesByIDs[instanceID] = internalInstance;
  }
}

function unregisterInstance(internalInstance) {
  var instanceID = getIDForInstance(internalInstance);
  if (instanceID) {
    delete instancesByIDs[instanceID];
  }
}

var ReactDebugInstanceMap = {
  getIDForInstance: function (internalInstance) {
    if (!checkValidInstance(internalInstance)) {
      return null;
    }
    return getIDForInstance(internalInstance);
  },
  getInstanceByID: function (instanceID) {
    return getInstanceByID(instanceID);
  },
  isRegisteredInstance: function (internalInstance) {
    if (!checkValidInstance(internalInstance)) {
      return false;
    }
    return isRegisteredInstance(internalInstance);
  },
  registerInstance: function (internalInstance) {
    if (!checkValidInstance(internalInstance)) {
      return;
    }
    process.env.NODE_ENV !== 'production' ? warning(!isRegisteredInstance(internalInstance), 'There is an internal error in the React developer tools integration. ' + 'A registered instance should not be registered again. ' + 'Please report this as a bug in React.') : void 0;
    registerInstance(internalInstance);
  },
  unregisterInstance: function (internalInstance) {
    if (!checkValidInstance(internalInstance)) {
      return;
    }
    process.env.NODE_ENV !== 'production' ? warning(isRegisteredInstance(internalInstance), 'There is an internal error in the React developer tools integration. ' + 'An unregistered instance should not be unregistered again. ' + 'Please report this as a bug in React.') : void 0;
    unregisterInstance(internalInstance);
  }
};

module.exports = ReactDebugInstanceMap;