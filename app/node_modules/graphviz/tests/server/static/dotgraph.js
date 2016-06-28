var DotGraphClient = {
  version: '0.0.1',
  host: 'http://localhost:3000',
  
  require: function(libraryName) {
    try{
      document.write('<script type="text/javascript" src="'+libraryName+'"><\/script>');
    } catch(e) {
      var script = document.createElement('script');
      script.type = 'text/javascript';
      script.src = libraryName;
      document.getElementsByTagName('head')[0].appendChild(script);
    }
  },
  
  load: function() {
    DotGraphClient.require( 'http://code.jquery.com/jquery-latest.min.js' );
  }
};

DotGraphClient.load();

var DotGraph = {
  file: function( url, id ) {
    $.get(DotGraphClient.host+'/file/'+url, function(data) {
      $(id).html(data);
    });
  },
  
  script: function( data, id ) {
    $.post(DotGraphClient.host+'/script', { data: data }, function(data) {
      $(id).html(data);
    });
  }
};
