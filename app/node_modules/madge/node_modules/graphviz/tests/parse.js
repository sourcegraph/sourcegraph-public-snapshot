var util = require('util'),
	graphviz = require('../lib/graphviz');

graphviz.parse( "cluster.dot", function(graph) {
	graph.render( "png", "cluster.png" );
})

graphviz.parse( "cluster.dot", function(graph) {
	graph.render( "png", function(render) {
		process.stdout.write( render );
	}, function(code, out, err) {
		console.log("RENDER ERROR")
		console.log(""+err)
		console.log(""+out)
	});
}, function(code, out, err) {
	console.log("PARSE ERROR")
	console.log(""+err)
	console.log(""+out)
});
