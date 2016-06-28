var util = require('util'),
	graphviz = require('../lib/graphviz');

// digraph G {
var g = graphviz.digraph("G");
// 	subgraph cluster_0 {
var cluster_0 = g.addCluster("cluster_0");
// 		style=filled;
cluster_0.set( "style", "filled" );
// 		color=lightgrey;
cluster_0.set( "color", "lightgrey" );
// 		node [style=filled,color=white];
cluster_0.setNodeAttribut( "style", "filled" )
cluster_0.setNodeAttribut( "color", "white" )

// 		a0 -> a1 -> a2 -> a3;
cluster_0.addEdge( "a0", "a1" );
cluster_0.addEdge( "a1", "a2" );
cluster_0.addEdge( "a2", "a3" );
// 		label = "process #1";
cluster_0.set( "label", "process #1" )
// 	}

// 	subgraph cluster_1 {
var cluster_1 = g.addCluster("cluster_1");
// 		node [style=filled];
cluster_1.setNodeAttribut( "style", "filled" )

// 		b0 -> b1 -> b2 -> b3;
cluster_1.addEdge( "b0", "b1" );
cluster_1.addEdge( "b1", "b2" );
cluster_1.addEdge( "b2", "b3" );
// 		label = "process #2";
cluster_1.set( "label", "process #2" );
// 		color = blue
cluster_1.set( "color", "blue" );
// 	}

// 	start -> a0;
g.addEdge( "start", cluster_0.getNode("a0") );
// 	start -> b0;
g.addEdge( "start", cluster_1.getNode("b0") );
// 	a1 -> b3;
g.addEdge( cluster_0.getNode("a1"), cluster_1.getNode("b3") );
// 	b2 -> a3;
g.addEdge( cluster_1.getNode("b2"), cluster_0.getNode("a3") );
// 	a3 -> a0;
g.addEdge( cluster_0.getNode("a3"), cluster_0.getNode("a0") );
// 	a3 -> end;
g.addEdge( cluster_0.getNode("a3"), "end" );
// 	b3 -> end;
g.addEdge( cluster_1.getNode("b3"), "end" );
// 
// 	start [shape=Mdiamond];
g.getNode("start").set( "shape", "Mdiamond" );
// 	end [shape=Msquare];
g.getNode("end").set( "shape", "Msquare" );
// }

g.output( "png", "cluster.png" ); 
