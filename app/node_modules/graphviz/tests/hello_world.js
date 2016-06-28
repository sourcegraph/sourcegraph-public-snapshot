var util = require('util'),
	graphviz = require('../lib/graphviz');

// Create digraph G
var g = graphviz.digraph("G");

// Add node (ID: Hello)
var n1 = g.addNode( "Hello", {"color" : "blue", "shape" : "Mdiamond"} );
//n1.set( "color", "blue" );
n1.set( "style", "filled" );

// Add node (ID: World)
g.addNode( "World" );

// Add edge between the two nodes
var e = g.addEdge( n1, "World" );
e.set( "color", "red" );

// Generate a PNG output
g.output( "png", "hello_world.png" );
