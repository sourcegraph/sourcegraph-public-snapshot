var assert = require('assert');
var matches = require('../lib/matches');

describe('matches', function() {

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


  it('works testing against a CSS selector', function() {
    fixtures.innerHTML = '<div id="foo" class="bar"></div>';
    var div = document.getElementById('foo');

    assert(matches(div, 'div'));
    assert(matches(div, '#foo'));
    assert(matches(div, 'body .bar'));

    assert(!matches(div, 'p'));
    assert(!matches(div, '#bar'));
  });


  it('works testing against a DOM element', function() {
    fixtures.innerHTML = '<div id="foo" class="bar"></div>';
    var div = document.getElementById('foo');

    assert(matches(div, fixtures.childNodes[0]));
    assert(!matches(div, fixtures));
  });


  it('works testing against a list of selectors and elements', function() {
    fixtures.innerHTML = '<div class="foo"><p id="bar"></p></div>';
    var p = document.getElementById('bar');

    assert(matches(p, ['#bar']));
    assert(matches(p, ['html', 'body', 'p']));
    assert(matches(p, ['#fixtures, p', document, document.body]));
    assert(matches(p, [fixtures, '.foo > #bar']));

    assert(!matches(p, ['html', 'body', fixtures]));
    assert(!matches(p, [document.body, 'span']));
  });


  it('handles invalid inputs gracefully', function() {
    assert(!matches());
    assert(!matches(fixtures, null));
    assert(!matches(document, '*'));
  });

});
