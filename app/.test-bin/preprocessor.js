var babel = require("babel-core");

module.exports = {
  process: function(src, filename) {
    if (filename.match(/\.css$/)) {
      return "";
    }

    // Ignore all files within node_modules
    // babel files can be .js, .es, .jsx or .es6
    if (filename.indexOf("node_modules") === -1 && babel.canCompile(filename)) {
      return babel.transform(src, {
        filename: filename,
        stage: process.env.BABEL_JEST_STAGE || 2,
        retainLines: true,
      }).code;
    }

    return src;
  }
};
