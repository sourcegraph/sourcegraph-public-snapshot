var util = require('util'),
  graphviz = require('../lib/graphviz');
  
// graph G {
var g = graphviz.graph("G");
//   e
var e = g.addNode( "e" )
//   subgraph clusterA {
var clusterA = g.addCluster( "clusterA" )
//     a -- b;
clusterA.addEdge( "a", "b" )
//     subgraph clusterC {
var clusterC = clusterA.addCluster( "clusterC" )
//       C -- D;
clusterC.addEdge( "C", "D" )
//     }
//   }
//   subgraph clusterB {
var clusterB = g.addCluster( "clusterB" )
clusterB.addEdge( "d", "f" )
//     d -- f
//   }
//   d -- D
g.addEdge( clusterB.getNode("d"), clusterC.getNode("D") )
//   e -- clusterB
g.addEdge( e, clusterB )
//   clusterC -- clusterB
g.addEdge( clusterC, clusterB )
// }

g.use = "fdp"
g.output( "png", "fdpclust.png" ); 
