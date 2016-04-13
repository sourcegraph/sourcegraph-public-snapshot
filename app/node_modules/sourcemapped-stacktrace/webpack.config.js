module.exports = {
  context: __dirname,
  entry: "./sourcemapped-stacktrace.js",
  output: {
    library: "sourceMappedStackTrace",
    libraryTarget: "umd",
    path: __dirname + "/dist",
    filename: "sourcemapped-stacktrace.js"
  }
};
