'use strict';

Object.defineProperty(exports, "__esModule", {
    value: true
});

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var React = require('react');
var IconBase = require('react-icon-base');

var TiThermometer = function (_React$Component) {
    _inherits(TiThermometer, _React$Component);

    function TiThermometer() {
        _classCallCheck(this, TiThermometer);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiThermometer).apply(this, arguments));
    }

    _createClass(TiThermometer, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.666666666666668 25.116666666666667v-9.283333333333333c0-0.4583333333333339-0.375-0.8333333333333339-0.8333333333333321-0.8333333333333339s-0.8333333333333321 0.375-0.8333333333333321 0.8333333333333339v9.283333333333333c-1.4333333333333336 0.375-2.5 1.6666666666666679-2.5 3.2166666666666686 0 1.8383333333333347 1.4933333333333323 3.333333333333332 3.333333333333332 3.333333333333332s3.333333333333332-1.495000000000001 3.333333333333332-3.333333333333332c0-1.5500000000000007-1.0666666666666664-2.8416666666666686-2.5-3.2166666666666686z m5-2.68333333333333v-13.26666666666667c0-3.2166666666666677-2.616666666666667-5.833333333333334-5.833333333333332-5.833333333333334s-5.833333333333336 2.6166666666666663-5.833333333333336 5.833333333333334v13.26666666666667c-1.536666666666667 1.5-2.5 3.583333333333332-2.5 5.899999999999999 0 4.594999999999999 3.7383333333333333 8.333333333333336 8.333333333333336 8.333333333333336s8.333333333333336-3.7383333333333297 8.333333333333336-8.333333333333336c0-2.3166666666666664-0.9633333333333347-4.399999999999999-2.5-5.899999999999999z m-5.833333333333332 10.899999999999999c-2.7566666666666677 0-5-2.2433333333333323-5-5 0-1.8416666666666686 1.0116666666666667-3.4366666666666674 2.5-4.305v-14.861666666666668c0-1.3783333333333339 1.1216666666666661-2.5 2.5-2.5s2.5 1.121666666666667 2.5 2.5v14.861666666666668c1.4883333333333333 0.8666666666666671 2.5 2.4633333333333347 2.5 4.305 0 2.7566666666666677-2.2433333333333323 5-5 5z' })
                )
            );
        }
    }]);

    return TiThermometer;
}(React.Component);

exports.default = TiThermometer;
module.exports = exports['default'];