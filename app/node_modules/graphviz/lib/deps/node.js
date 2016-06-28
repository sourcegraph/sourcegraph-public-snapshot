/**
 * Module dependencies
 */
var Hash = require('./core_ext/hash').Hash,
  Attributs = require('./attributs').Attributs;

/**
 * Create a new node
 * @constructor
 * @param {Graph} graph Parent Graph
 * @param {String} id The node ID
 * @return {Node}
 * @api public
 */
var Node = exports.Node = function(graph, id) {
  this.relativeGraph = graph;
  this.id = id;
  this.attributs = new Attributs("N");
};

/**
 *
 */
Node.prototype.to = function(id, attrs) {
	this.relativeGraph.addEdge(this, id, attrs);
	return this.relativeGraph.from(id);
};

/**
 * Set a node attribut
 *
 * @param {String} name The attribut name
 * @param {Void} value The attribut value
 * @api public
 */
Node.prototype.set = function(name, value) {
  this.attributs.set(name, value);
};

/**
 * Get a node attribut
 *
 * @param {String} name The attribut name
 * @return {Void}
 * @api public
 */
Node.prototype.get = function(name) {
  return this.attributs.get(name);
};

/**
 * @api private
 */
Node.prototype.to_dot = function() {
  var nodeOutput = '"' + this.id + '"' + this.attributs.to_dot();
  return nodeOutput;
};
