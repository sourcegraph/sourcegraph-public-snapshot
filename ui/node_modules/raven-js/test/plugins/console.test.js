var _Raven = require('../../src/raven');
var consolePlugin = require('../../plugins/console');

var Raven;
describe('console plugin', function () {
    beforeEach(function () {
        Raven = new _Raven();
        Raven.config('http://abc@example.com:80/2');
    });

    it('should call Raven.captureMessage', function () {
        var console = {
            debug: function () {},
            info: function () {},
            warn: function () {},
            error: function () {}
        };

        consolePlugin(Raven, console);

        this.sinon.stub(Raven, 'captureMessage');
        console.error('Raven should capture', 'console.error');

        assert.equal(Raven.captureMessage.callCount, 1);
        assert.equal(Raven.captureMessage.getCall(0).args[0], 'Raven should capture console.error');
        assert.deepEqual(Raven.captureMessage.getCall(0).args[1], {
            level: 'error',
            logger: 'console',
            extra: {
                arguments: ['Raven should capture', 'console.error']
            }
        });

        Raven.captureMessage.reset();

        console.warn('Raven should capture console.warn');

        assert.equal(Raven.captureMessage.callCount, 1);
        assert.equal(Raven.captureMessage.getCall(0).args[0], 'Raven should capture console.warn');
        assert.deepEqual(Raven.captureMessage.getCall(0).args[1], {
            level: 'warning', // warn => warning
            logger: 'console',
            extra: {
                arguments: ['Raven should capture console.warn']
            }
        });
    });
});
