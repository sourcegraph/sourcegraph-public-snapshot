var type = require('./type');
var utils = require('./utils');

/*
 * Wrapper for a user properties JSON object that supports operations.
 * Note: if a user property is used in multiple operations on the same Identify object,
 * only the first operation will be saved, and the rest will be ignored.
 */

var AMP_OP_ADD = '$add';
var AMP_OP_APPEND = '$append';
var AMP_OP_CLEAR_ALL = '$clearAll';
var AMP_OP_PREPEND = '$prepend';
var AMP_OP_SET = '$set';
var AMP_OP_SET_ONCE = '$setOnce';
var AMP_OP_UNSET = '$unset';

var Identify = function() {
  this.userPropertiesOperations = {};
  this.properties = []; // keep track of keys that have been added
};

Identify.prototype.add = function(property, value) {
  if (type(value) === 'number' || type(value) === 'string') {
    this._addOperation(AMP_OP_ADD, property, value);
  } else {
    utils.log('Unsupported type for value: ' + type(value) + ', expecting number or string');
  }
  return this;
};

Identify.prototype.append = function(property, value) {
  this._addOperation(AMP_OP_APPEND, property, value);
  return this;
};

// clearAll should be sent on its own Identify object
// If there are already other operations, then don't add clearAll
// If clearAll already in Identify, don't add other operations
Identify.prototype.clearAll = function() {
  if (Object.keys(this.userPropertiesOperations).length > 0) {
    if (!this.userPropertiesOperations.hasOwnProperty(AMP_OP_CLEAR_ALL)) {
      utils.log('Need to send $clearAll on its own Identify object without any other operations, skipping $clearAll');
    }
    return this;
  }
  this.userPropertiesOperations[AMP_OP_CLEAR_ALL] = '-';
  return this;
};

Identify.prototype.prepend = function(property, value) {
  this._addOperation(AMP_OP_PREPEND, property, value);
  return this;
};

Identify.prototype.set = function(property, value) {
  this._addOperation(AMP_OP_SET, property, value);
  return this;
};

Identify.prototype.setOnce = function(property, value) {
  this._addOperation(AMP_OP_SET_ONCE, property, value);
  return this;
};

Identify.prototype.unset = function(property) {
  this._addOperation(AMP_OP_UNSET, property, '-');
  return this;
};

Identify.prototype._addOperation = function(operation, property, value) {
  // check that the identify doesn't already contain a clearAll
  if (this.userPropertiesOperations.hasOwnProperty(AMP_OP_CLEAR_ALL)) {
    utils.log('This identify already contains a $clearAll operation, skipping operation ' + operation);
    return;
  }

  // check that property wasn't already used in this Identify
  if (this.properties.indexOf(property) !== -1) {
    utils.log('User property "' + property + '" already used in this identify, skipping operation ' + operation);
    return;
  }

  if (!this.userPropertiesOperations.hasOwnProperty(operation)){
    this.userPropertiesOperations[operation] = {};
  }
  this.userPropertiesOperations[operation][property] = value;
  this.properties.push(property);
};

module.exports = Identify;
