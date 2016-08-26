var assert = require('assert');
var parents = require('../lib/parents');

describe('parents', function() {

  it('returns an array of a DOM elements parent elements', function() {
    var fixtures = document.createElement('div');
    fixtures.id = 'fixtures';
    document.body.appendChild(fixtures);

    fixtures.innerHTML = '<p id="p"><em id="em"><sup id="sup"></sup><em></p>';
    var p = document.getElementById('p');
    var em = document.getElementById('em');
    var sup = document.getElementById('sup');

    assert.deepEqual(parents(sup), [
      em,
      p,
      fixtures,
      document.body,
      document.documentElement
    ]);

    document.body.removeChild(fixtures);
  });


  it('returns an empty array if no parents exist', function() {
    assert.deepEqual(parents(document.documentElement), []);
    assert.deepEqual(parents(document.createElement('div')), []);
  });


  it('handles invalid input gracefully', function() {
    assert.deepEqual(parents(), []);
    assert.deepEqual(parents(null), []);
    assert.deepEqual(parents(document), []);
  });

});

