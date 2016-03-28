'use strict';

var gulp = require('gulp'),
clean = require('gulp-clean'),
cleanhtml = require('gulp-cleanhtml'),
minifycss = require('gulp-minify-css'),
jshint = require('gulp-jshint'),
template = require('gulp-template'),
sass = require('gulp-sass'),
stripdebug = require('gulp-strip-debug'),
uglify = require('gulp-uglify'),
zip = require('gulp-zip');

var context = {
  DEV: process.env.DEV || false,

  /*global process*/
  url: (process.env.DEV ? 'http://localhost:3080' : 'https://sourcegraph.com'),
};

// Clean build directory
gulp.task('clean', function() {
  return gulp.src('build/*', {read: false})
    .pipe(clean());
});

// Copy static folders to build directory
gulp.task('copy', function() {
  gulp.src('manifest.json')
    .pipe(template(context))
    .pipe(gulp.dest('build'));
  return gulp.src(['*.png']).pipe(gulp.dest('build'));
});

// Copy and compress HTML files
gulp.task('html', function() {
  return gulp.src('*.html')
    .pipe(template(context))
    .pipe(cleanhtml())
    .pipe(gulp.dest('build'));
});

// Run scripts through JSHint
gulp.task('jshint', function() {
  return gulp.src(['*.js', '!gulpfile.js'])
    .pipe(jshint())
    .pipe(jshint.reporter('default'));
});

// Copy vendor scripts and uglify all other scripts, creating source maps
// gulp.task('scripts', ['jshint'], function() {
gulp.task('scripts', function() {
  var s = gulp.src(['*.js', '!gulpfile.js'])
    .pipe(template(context));
  if (!context.DEV) {
    s = s.pipe(uglify());
  }
  return s.pipe(gulp.dest('build'));
});

// Minify styles
gulp.task('styles', function() {
  return gulp.src('*.scss')
    .pipe(sass({errLogToConsole: true}))
    .pipe(minifycss({root: '.', keepSpecialComments: 0}))
    .pipe(gulp.dest('build'));
});

gulp.task('build', ['clean', 'html', 'scripts', 'styles', 'copy']);

gulp.task('watch', function() {
  gulp.watch(['*.js', '*.scss', '*.html', '*.json'], ['build']);
});

// Run all tasks after build directory has been cleaned
gulp.task('default', ['clean', 'build', 'watch']);
