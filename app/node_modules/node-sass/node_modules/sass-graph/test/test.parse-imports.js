var parseImports = require('../parse-imports');
var assert = require("assert");
var path = require("path");

describe('parse-imports', function () {

  it('should parse single import with single quotes', function () {
    var scss = "@import 'app'; ";
    var result = parseImports(scss);
    assert.equal(result.length, 1);
    assert.equal(result[0], "app");
  });

  it('should parse single import with double quotes', function () {
    var scss = '@import "app"; ';
    var result = parseImports(scss);
    assert.equal(result.length, 1);
    assert.equal(result[0], "app");
  });

  it('should parse single import with extra spaces after import', function () {
    var scss = '@import  "app"; ';
    var result = parseImports(scss);
    assert.equal(result.length, 1);
    assert.equal(result[0], "app");
  });

  it('should parse single import with extra spaces before ;', function () {
    var scss = '@import "app" ; ';
  });

  it('should parse two individual imports', function () {
    var scss = '@import "app"; \n ' + 
               '@import "foo"; \n';
    var result = parseImports(scss);
    ["app", "foo"].forEach(function (dep) {
      assert.equal(result.length, 2);
      assert.notEqual(result.indexOf(dep), -1);
    });       
  });

  it('should parse two imports on same line', function () {
    var scss = '@import "app", "foo";';  
    var result = parseImports(scss);
    ["app", "foo"].forEach(function (dep) {
      assert.equal(result.length, 2);
      assert.notEqual(result.indexOf(dep), -1);
    });       
  });

  it('should parse two imports continued on multiple lines ', function () {
    var scss = '@import "app", \n' +
                   '"foo"; \n';  
    var result = parseImports(scss);               
    ["app", "foo"].forEach(function (dep) {
      assert.equal(result.length, 2);
      assert.notEqual(result.indexOf(dep), -1);
    });       
  });

  it('should parse three imports with mixed style ', function () {
    var scss = '@import "app", \n' +
                        '"foo";\n ' +  
               '@import "bar";';
    var result = parseImports(scss);
    ["app", "foo", "bar"].forEach(function (dep) {
      assert.equal(result.length, 3);
      assert.notEqual(result.indexOf(dep), -1);
    });       
  });


});