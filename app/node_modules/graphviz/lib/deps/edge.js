/**
 * Module dependencies
 */
var Hash = require('./core_ext/hash').Hash,
  Attributs = require('./attributs').Attributs;

/**
 * Create a new edge
 * @constructor
 * @param {Graph} graph Parent Graph
 * @param {String|Node} nodeOne The first node
 * @param {String|Node} nodeTwo The second node
 * @return {Edge}
 * @api public
 */
var Edge = exports.Edge = function(graph, nodeOne, nodeTwo) {
  this.relativeGraph = graph;
  this.nodeOne = nodeOne;
  this.nodeTwo = nodeTwo;
  this.attributs = new Attributs("E");
};

/**
 * Set an edge attribut
 *
 * @param {String} name The attribut name
 * @param {Void} value The attribut value
 * @api public
 */
Edge.prototype.set = function(name, value) {
  this.attributs.set(name, value);
};

/**
 * Get an edge attribut
 *
 * @param {String} name The attribut name
 * @return {Void}
 * @api public
 */
Edge.prototype.get = function(name) {
  return this.attributs.get(name);
};

/**
 * @api private
 */
Edge.prototype.to_dot = function() {
  var edgeLink = "->";
  if(this.relativeGraph.type === "graph") {
    edgeLink = "--";
  }
  
  var edgeOutput = '"' + this.nodeOne.id + '"' + " " + edgeLink + " " + '"' + this.nodeTwo.id + '"';
  edgeOutput = edgeOutput + this.attributs.to_dot();
  return edgeOutput;
};
