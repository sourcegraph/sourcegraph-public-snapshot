var _Raven = require('../../src/raven');
var angularPlugin = require('../../plugins/angular');

var Raven;
describe('Angular plugin', function () {
    beforeEach(function () {
        Raven = new _Raven();
        Raven.config('http://abc@example.com:80/2');
    });

    describe('_normalizeData()', function () {
        it('should extract type, value, and the angularDocs URL from unminified exceptions', function () {
            var data = {
                project: '2',
                logger: 'javascript',
                platform: 'javascript',

                culprit: 'http://example.org/app.js',
                message: 'Error: crap',
                exception: {
                    type: 'Error',
                    values: [{
                        value:
                          '[ngRepeat:dupes] Duplicates in a repeater are not allowed. Use \'track by\' expression to specify unique keys. Repeater: element in elements | orderBy: \'createdAt\' track by element.id, Duplicate key: 25, Duplicate value: {"id":"booking-40"}\n' +
                          'http://errors.angularjs.org/1.4.3/ngRepeat/dupes?p0=element%20in%20elements…00%3A00%22%2C%22ends_at%22%3A%222016-06-09T23%3A59%3A59%2B00%3A00%22%7D%7D'
                    }]
                },
                extra: {}
            };

            angularPlugin._normalizeData(data);

            var exception = data.exception.values[0];
            assert.equal(exception.type, 'ngRepeat:dupes');
            assert.equal(exception.value, 'Duplicates in a repeater are not allowed. Use \'track by\' expression to specify unique keys. Repeater: element in elements | orderBy: \'createdAt\' track by element.id, Duplicate key: 25, Duplicate value: {"id":"booking-40"}');
            assert.equal(data.extra.angularDocs, 'http://errors.angularjs.org/1.4.3/ngRepeat/dupes?p0=element%20in%20elements…00%3A00%22%2C%22ends_at%22%3A%222016-06-09T23%3A59%3A59%2B00%3A00%22%7D%7D');
        });

        it('should extract type and the angularDocs URL minified exceptions', function () {
            var data = {
                project: '2',
                logger: 'javascript',
                platform: 'javascript',

                culprit: 'http://example.org/app.js',
                message: 'Error: crap',
                exception: {
                    type: 'Error',
                    values: [{
                        // no message, no newlines
                        value: '[ngRepeat:dupes] http://errors.angularjs.org/1.4.3/ngRepeat/dupes?p0=element%20in%20elements…00%3A00%22%2C%22ends_at%22%3A%222016-06-09T23%3A59%3A59%2B00%3A00%22%7D%7D'
                    }]
                },
                extra: {}
            };

            angularPlugin._normalizeData(data);

            var exception = data.exception.values[0];
            assert.equal(exception.type, 'ngRepeat:dupes');
            assert.equal(exception.value, '');
            assert.equal(data.extra.angularDocs, 'http://errors.angularjs.org/1.4.3/ngRepeat/dupes?p0=element%20in%20elements…00%3A00%22%2C%22ends_at%22%3A%222016-06-09T23%3A59%3A59%2B00%3A00%22%7D%7D');
        });
    });
});
