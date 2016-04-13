describe('getUtmData', function() {
  var getUtmData = require('../src/utm.js');

  it('should get utm params from the query string', function() {
    var query = '?utm_source=amplitude&utm_medium=email&utm_term=terms' +
                '&utm_content=top&utm_campaign=new';
    var utms = getUtmData('', query);
    assert.deepEqual(utms, {
      utm_campaign: 'new',
      utm_content: 'top',
      utm_medium: 'email',
      utm_source: 'amplitude',
      utm_term: 'terms'
    });
  });

  it('should get utm params from the cookie string', function() {
    var cookie = '133232535.1424926227.1.1.utmcsr=google|utmccn=(organic)' +
                 '|utmcmd=organic|utmctr=(none)|utmcct=link';
    var utms = getUtmData(cookie, '');
    assert.deepEqual(utms, {
      utm_campaign: '(organic)',
      utm_content: 'link',
      utm_medium: 'organic',
      utm_source: 'google',
      utm_term: '(none)'
    });
  });

  it('should prefer utm params from the query string', function() {
    var query = '?utm_source=amplitude&utm_medium=email&utm_term=terms' +
                '&utm_content=top&utm_campaign=new';
    var cookie = '133232535.1424926227.1.1.utmcsr=google|utmccn=(organic)' +
                 '|utmcmd=organic|utmctr=(none)|utmcct=link';
    var utms = getUtmData(cookie, query);
    assert.deepEqual(utms, {
      utm_campaign: 'new',
      utm_content: 'top',
      utm_medium: 'email',
      utm_source: 'amplitude',
      utm_term: 'terms'
    });
  });
});
