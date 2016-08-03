'use strict';

exports.__esModule = true;

var _createClass = (function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ('value' in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; })();

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

function _inherits(subClass, superClass) { if (typeof superClass !== 'function' && superClass !== null) { throw new TypeError('Super expression must either be null or a function, not ' + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var _container = require('./container');

var _container2 = _interopRequireDefault(_container);

var _warnOnce = require('./warn-once');

var _warnOnce2 = _interopRequireDefault(_warnOnce);

var AtRule = (function (_Container) {
    _inherits(AtRule, _Container);

    function AtRule(defaults) {
        _classCallCheck(this, AtRule);

        _Container.call(this, defaults);
        this.type = 'atrule';
    }

    AtRule.prototype.append = function append() {
        var _Container$prototype$append;

        if (!this.nodes) this.nodes = [];

        for (var _len = arguments.length, children = Array(_len), _key = 0; _key < _len; _key++) {
            children[_key] = arguments[_key];
        }

        return (_Container$prototype$append = _Container.prototype.append).call.apply(_Container$prototype$append, [this].concat(children));
    };

    AtRule.prototype.prepend = function prepend() {
        var _Container$prototype$prepend;

        if (!this.nodes) this.nodes = [];

        for (var _len2 = arguments.length, children = Array(_len2), _key2 = 0; _key2 < _len2; _key2++) {
            children[_key2] = arguments[_key2];
        }

        return (_Container$prototype$prepend = _Container.prototype.prepend).call.apply(_Container$prototype$prepend, [this].concat(children));
    };

    AtRule.prototype.insertBefore = function insertBefore(exist, add) {
        if (!this.nodes) this.nodes = [];
        return _Container.prototype.insertBefore.call(this, exist, add);
    };

    AtRule.prototype.insertAfter = function insertAfter(exist, add) {
        if (!this.nodes) this.nodes = [];
        return _Container.prototype.insertAfter.call(this, exist, add);
    };

    _createClass(AtRule, [{
        key: 'afterName',
        get: function get() {
            _warnOnce2['default']('AtRule#afterName was deprecated. Use AtRule#raws.afterName');
            return this.raws.afterName;
        },
        set: function set(val) {
            _warnOnce2['default']('AtRule#afterName was deprecated. Use AtRule#raws.afterName');
            this.raws.afterName = val;
        }
    }, {
        key: '_params',
        get: function get() {
            _warnOnce2['default']('AtRule#_params was deprecated. Use AtRule#raws.params');
            return this.raws.params;
        },
        set: function set(val) {
            _warnOnce2['default']('AtRule#_params was deprecated. Use AtRule#raws.params');
            this.raws.params = val;
        }
    }]);

    return AtRule;
})(_container2['default']);

exports['default'] = AtRule;
module.exports = exports['default'];