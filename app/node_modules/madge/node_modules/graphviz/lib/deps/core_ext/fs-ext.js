var fs = require('fs');

exports.exist = function(path) {
  try {
    fs.statSync( path );
  }
  catch(e) {
    return( false );
  }
  return( true );
}
