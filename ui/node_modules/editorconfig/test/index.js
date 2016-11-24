var editorconfig = require('../');
var fs = require('fs');
var path = require('path');
var should = require('should');

describe('parse', function() {
  it('async', function() {
    var expected = {
      indent_style: 'space',
      indent_size: 2,
      end_of_line: 'lf',
      charset: 'utf-8',
      trim_trailing_whitespace: true,
      insert_final_newline: true,
      tab_width: 2
    };
    var target = path.join(__dirname, '/app.js');
    var promise = editorconfig.parse(target);
    return promise.then(function onFulfilled(result) {
      expected.should.eql(result);
    });
  });

  it('sync', function() {
    var expected = {
      indent_style: 'space',
      indent_size: 2,
      end_of_line: 'lf',
      charset: 'utf-8',
      trim_trailing_whitespace: true,
      insert_final_newline: true,
      tab_width: 2
    };
    var target = path.join(__dirname, '/app.js');
    expected.should.eql(editorconfig.parseSync(target));
  });
});

describe('parseFromFiles', function() {
  it('async', function() {
    var expected = {
      indent_style: 'space',
      indent_size: 2,
      tab_width: 2,
      end_of_line: 'lf',
      charset: 'utf-8',
      trim_trailing_whitespace: true,
      insert_final_newline: true,
    };
    var configs = [];
    var configPath = path.resolve(__dirname, '../.editorconfig');
    var config = {
      name: configPath,
      contents: fs.readFileSync(configPath, 'utf8')
    };
    configs.push(config);
    var target = path.join(__dirname, '/app.js');
    var promise = editorconfig.parseFromFiles(target, configs);
    return promise.then(function onFulfilled(result) {
      expected.should.eql(result);
    });
  });

  it('sync', function() {
    var expected = {
      indent_style: 'space',
      indent_size: 2,
      tab_width: 2,
      end_of_line: 'lf',
      charset: 'utf-8',
      trim_trailing_whitespace: true,
      insert_final_newline: true,
    };
    var configs = [];
    var configPath = path.resolve(__dirname, '../.editorconfig');
    var config = {
      name: configPath,
      contents: fs.readFileSync(configPath, 'utf8')
    };
    configs.push(config);
    var target = path.join(__dirname, '/app.js');
    expected.should.eql(editorconfig.parseFromFilesSync(target, configs));
  });
});
