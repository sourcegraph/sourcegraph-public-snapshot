var util = require('util'),
	graphviz = require('../lib/graphviz');

// Create digraph G
var g = graphviz.digraph("G");

// Add node (ID: Hello)
var n1 = g.addNode( "Hello" );

// Add node (ID: World)
g.addNode( "World" );

// Add edge between the two nodes
var e = g.addEdge( n1, "World" );

// Generate a PNG output
g.output( {
	"type":"png",
	"use":"dot",
	"N" : {
		"color":"blue",
		"shape":"Mdiamond"
	},
	"E" : {
		"color" : "red",
		"label" : "Say"
	},
	"G" : {
		"label" : "Example"
	}
}, "hello.png" );
