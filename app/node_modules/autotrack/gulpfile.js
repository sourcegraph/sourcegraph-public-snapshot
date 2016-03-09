/**
 * Copyright 2016 Google Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */


var browserify = require('browserify');
var buffer = require('vinyl-buffer');
var connect = require('connect');
var fs = require('fs');
var gulp = require('gulp');
var gutil = require('gulp-util');
var seleniumServerJar = require('selenium-server-standalone-jar');
var pkg = require('./package.json');
var shell = require('shelljs');
var serveStatic = require('serve-static');
var source = require('vinyl-source-stream');
var sourcemaps = require('gulp-sourcemaps');
var spawn = require('child_process').spawn;
var uglify = require('gulp-uglify');
var webdriver = require('gulp-webdriver');


var server;


gulp.task('javascript', function(done) {

  // Gets the license string from this file, the first 15 lines.
  var license = fs.readFileSync(__filename, 'utf-8')
      .split('\n').slice(0, 15).join('\n');

  var version = '/*! autotrack.js v' + pkg.version + ' */';

  browserify('./', {
    debug: true
  })
  .bundle()

  // TODO(philipwalton): Add real error handling.
  // This temporary hack fixes an issue with tasks not restarting in
  // watch mode after a syntax error is fixed.
  .on('error', function(err) { gutil.beep(); done(err); })
  .on('end', done)

  .pipe(source('./autotrack.js'))
  .pipe(buffer())
  .pipe(sourcemaps.init({loadMaps: true}))
  .pipe(uglify({output: {preamble: license + '\n\n' + version}}))
  .pipe(sourcemaps.write('./'))
  .pipe(gulp.dest('./'));
});


gulp.task('test', ['javascript', 'serve', 'selenium'], function() {
  function stopServers() {
    server.close();
    if (!process.env.CI) seleniumServer.kill();
  }
  return gulp.src('./wdio.conf.js')
      .pipe(webdriver())
      .on('end', stopServers);
});


gulp.task('serve', ['javascript'], function(done) {
  server = connect().use(serveStatic('./')).listen(8080, done);
});


gulp.task('selenium', function(done) {
  // Don't start the selenium server on CI.
  if (process.env.CI) return done();

  seleniumServer = spawn('java',  ['-jar', seleniumServerJar.path]);
  seleniumServer.stderr.on('data', function(data) {
    if (data.indexOf('Selenium Server is up and running') > -1) {
      done();
    }
  });
});


gulp.task('watch', ['serve'], function() {
  gulp.watch('./lib/**/*.js', ['javascript']);
});


gulp.task('build', ['test']);
