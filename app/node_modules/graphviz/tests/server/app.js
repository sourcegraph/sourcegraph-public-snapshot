var fs = require('fs');
var util = require('util');
var http = require('http');
var url = require('url');

// npm install express
var express = require('express'),
  app = express.createServer();
// nom install temp
var temp = require('temp');
// npm install graphviz
var graphviz = require('../../lib/graphviz');
// npm install jsmin
var jsmin = require('jsmin').jsmin;

// Configuration --------------------------------------------------------------

app.set('views', __dirname + '/views');
app.configure(function(){ 
  app.use(express.bodyParser());
  app.use(express.static(__dirname + '/static'));
})

// Site -----------------------------------------------------------------------

app.get('/', function(req, res){
    res.render('index.ejs', {});
});

// APIs -----------------------------------------------------------------------

app.get('/dotgraph.min.js', function(req,res) {
  fs.readFile(__dirname+'/static/dotgraph.js', function (err, data) {
    if (err) throw err;
    res.contentType('text/javascript');
    res.send( jsmin(data.toString('utf8') ) );
  });
})

function __do( req, res, data ) {
  graphviz.parse( data, function(graph) {
    graph.render( "png", function(render) {
      img = '<img src="data:image/png;base64,'+render.toString("base64")+'"/>'
      res.send(img)
    }, function(code, out, err) {
      img = '<div class="error"><p><b>Render error (code '+code+')</b></p>';
      img += '<p>STDOUT : '+out+'</p>';
      img += '<p>STDERR : '+err+'</p></div>';
      res.send(img)
    });
  }, function(code, out, err){
    img = '<div class="error"><p><b>Parser error (code '+code+')</b></p>';
    img += '<p>STDERR : '+err+'</p></div>';
    img += '<p>STDOUT : '+out+'</p></div>';
    res.send(img)
  });  
}

app.post('/script', function(req,res){
  __do(req, res, req.body.data)
})

app.get('/file/*', function(req,res){
  var urlData = url.parse(req.params[0]);
  
  var urlPort = urlData.port;
  if( urlPort == undefined ) {
    urlPort = 80;
  }
  var urlHost = urlData.host;
  var urlPath = urlData.pathname;
  
  var client = http.createClient(urlPort, urlHost);
  var request = client.request('GET', urlPath,
    {'host': urlHost});
  request.end();
  request.on('response', function (response) {
    response.setEncoding('utf8');
    if(response.statusCode == 404 ) {
      res.send('<div class="error"><p><b>'+req.params[0]+'</b> does not exist (404 error)</p></div>');
    } else {
      response.on('data', function (chunk) {
        __do(req, res, chunk)
      });
    }
  });
})

app.listen(3000);
