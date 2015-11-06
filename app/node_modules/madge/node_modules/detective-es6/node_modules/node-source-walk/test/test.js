var assert = require('assert');
var Walker = require('../');
var fs = require('fs');
var sinon = require('sinon');

describe('node-source-walk', function() {
  var walker, ast, src;

  beforeEach(function() {
    walker = new Walker();
    src = fs.readFileSync(__dirname + '/example/srcFile.js', 'utf8');
    ast = walker.parse(src);
  });

  it('does not fail on binary scripts with a hashbang', function() {
    var walker = new Walker();
    var src = fs.readFileSync(__dirname + '/example/hashbang.js', 'utf8');
    assert.doesNotThrow(function() {
      var ast = walker.parse(src);
    });
  });

  it('parses es6 by default', function() {
    var walker = new Walker();
    assert.doesNotThrow(function() {
      walker.walk('() => console.log("foo")', function() {});
      walker.walk('import {foo} from "bar";', function() {});
    });
  });

  describe('walk', function() {
    var parseSpy, cb;

    beforeEach(function() {
      parseSpy = sinon.stub(walker, 'parse');
      cb = sinon.spy();
    });

    afterEach(function() {
      parseSpy.restore();
      cb.reset();
    });

    it('parses the given source code', function() {
      walker.walk(src, cb);

      assert(parseSpy.called);
    });

    it('calls the given callback for each node in the ast', function() {
      walker.walk(ast, cb);
      assert(cb.called);
      var node = cb.getCall(0).args[0];
      assert(typeof node === 'object');
    });

    it('reuses a given AST instead of parsing again', function() {
      walker.walk(ast, cb);

      assert.ok(!parseSpy.called);
      assert.ok(cb.called);
    });
  });

  describe('traverse', function() {
    it ('creates a parent reference for each node', function() {
      var cb = sinon.spy();
      walker.walk(ast, cb);
      var firstNode = cb.getCall(0).args[0];
      var secondNode = cb.getCall(1).args[0];
      assert(secondNode.parent === firstNode.body);
    });
  });

  describe('stopWalking', function() {
    it('halts further traversal of the AST', function() {
      var spy = sinon.spy();

      walker.walk(ast, function() {
        spy();
        walker.stopWalking();
      })

      assert(spy.calledOnce);
    });
  });
});
