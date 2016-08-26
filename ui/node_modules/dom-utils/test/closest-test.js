var assert = require('assert');
var closest = require('../lib/closest');

describe('closest', function() {

  var fixtures = document.createElement('div');
  fixtures.id = 'fixtures';


  beforeEach(function() {
    document.body.appendChild(fixtures);
  });


  afterEach(function() {
    fixtures.innerHTML = '';
  });


  after(function() {
    document.body.removeChild(fixtures);
  });


  it('should find a matching parent from a CSS selector', function() {
    fixtures.innerHTML =
        '<div id="div">' +
        '  <p id="p">' +
        '    <em id="em"></em>' +
        '  </p>' +
        '</div>';

    var div = document.getElementById('div');
    var p = document.getElementById('p');
    var em = document.getElementById('em');

    assert.equal(closest(em, 'p'), p);
    assert.equal(closest(em, '#div'), div);
    assert.equal(closest(p, 'html > body'), document.body);

    assert(!closest(em, '#nomatch'));
  });


  it('should test the element itself if the third args is true', function() {
    fixtures.innerHTML = '<p id="p"><em id="em"></em></p>';

    var p = document.getElementById('p');
    var em = document.getElementById('em');

    assert(!closest(em, 'em'));
    assert.equal(closest(em, 'em', true), em);
    assert(!closest(p, 'p'));
    assert.equal(closest(p, 'p', true), p);
  });


  it('handles invalid inputs gracefully', function() {
    assert(!closest());
    assert(!closest(null, 'div'));
    assert(!closest(document.body));
    assert(!closest(document, '*'));
  });

});
