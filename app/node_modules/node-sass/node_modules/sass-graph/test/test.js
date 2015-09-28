
var assert = require("assert");
var path = require("path");

var fixtures = path.resolve("test/fixtures");
var files = {
  'a.scss': fixtures + '/a.scss',
  'b.scss': fixtures + '/b.scss',
  '_c.scss': fixtures + '/_c.scss',
  'd.scss': fixtures + '/d.scss',
  '_e.scss': fixtures + '/components/_e.scss',
  'f.scss': fixtures + '/f.scss',
  'g.scss': fixtures + '/g.scss',
  '_h.scss': fixtures + '/nested/_h.scss',
  '_i.scss': fixtures + '/nested/_i.scss',
  'i.scss': fixtures + '/_i.scss',
  'j.scss': fixtures + '/j.scss',
  'k.l.scss': fixtures + '/components/k.l.scss',
  'm.scss': fixtures + '/m.scss',
  '_n.scss': fixtures + '/compass/_n.scss',
  '_compass.scss': fixtures + '/components/_compass.scss'
}

describe('sass-graph', function(){
  var sassGraph = require('../sass-graph');

  describe('parsing a graph of all scss files', function(){
    var graph = sassGraph.parseDir(fixtures, {loadPaths: [fixtures + '/components']});

    it('should have all files', function(){
      assert.equal(Object.keys(files).length, Object.keys(graph.index).length);
    })

    it('should have the correct imports for a.scss', function() {
      assert.deepEqual([files['b.scss']], graph.index[files['a.scss']].imports);
    });

    it('should have the correct importedBy for _c.scss', function() {
      assert.deepEqual([files['b.scss']], graph.index[files['_c.scss']].importedBy);
    });

    it('should have the correct (nested) imports for g.scss', function() {
      var expectedDescendents = [files['_h.scss'], files['_i.scss']];
      var descendents = [];

      graph.visitDescendents(files['g.scss'], function (imp) {
        descendents.push(imp);
        assert.notEqual(expectedDescendents.indexOf(imp), -1);
      });
    });

    it('should ignore custom imports for m.scss', function() {
      assert.deepEqual([files['_compass.scss'] , files['_n.scss']], graph.index[files['m.scss']].imports);
    });

    it('should traverse ancestors of _c.scss', function() {
      ancestors = [];
      graph.visitAncestors(files['_c.scss'], function(k) {
        ancestors.push(k);
      })
      assert.deepEqual([files['b.scss'], files['a.scss']], ancestors);
    });

    it('should prioritize cwd', function() {
      var expectedDescendents = [files['_i.scss']];
      var descendents = [];

      graph.visitDescendents(files['_h.scss'], function (imp) {
        descendents.push(imp);
        assert.notEqual(expectedDescendents.indexOf(imp), -1);
      });
    });
  })

  describe('parseFile', function () {
    it('should parse imports', function () {
      var graph = sassGraph.parseFile(files['a.scss']);
      var expectedDescendents = [files['b.scss'], files['_c.scss']];
      var descendents = [];
      graph.visitDescendents(files['a.scss'], function (imp) {
        descendents.push(imp);
        assert.notEqual(expectedDescendents.indexOf(imp), -1);
      });
      assert.equal(expectedDescendents.length, descendents.length);
    });
  });

  describe('parseFile', function () {
    it('should parse imports with loadPaths', function () {
      var graph = sassGraph.parseFile(files['d.scss'], {loadPaths: [fixtures + '/components']} );
      var expectedDescendents = [files['_e.scss']];
      var descendents = [];
      graph.visitDescendents(files['d.scss'], function (imp) {
        descendents.push(imp);
        assert.notEqual(expectedDescendents.indexOf(imp), -1);
      });
      assert.equal(expectedDescendents.length, descendents.length);
    });
  });

  describe('parseFile', function () {
    it('should thow an error', function () {
      try {
        var graph = sassGraph.parseFile(files['d.scss']);
      } catch (e) {
        assert.equal(e, "File to import not found or unreadable: e");
      }
    });
  });

  describe('parseFile', function () {
    it('should not throw an error for a file with no dependencies with Array having added functions', function () {
      try {
        Array.prototype.foo = function() {
          return false;
        }
        var graph = sassGraph.parseFile(files['f.scss']);
      } catch (e) {
        assert.fail("Error: " + e);
      }
    });
  });
});
