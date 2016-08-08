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

var TiDocumentText = function (_React$Component) {
    _inherits(TiDocumentText, _React$Component);

    function TiDocumentText() {
        _classCallCheck(this, TiDocumentText);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiDocumentText).apply(this, arguments));
    }

    _createClass(TiDocumentText, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.333333333333336 35h-16.666666666666668c-2.7566666666666677 0-5-2.2433333333333323-5-5v-20c0-2.756666666666667 2.243333333333334-5 5-5h16.666666666666668c2.7566666666666677 0 5 2.243333333333334 5 5v20c0 2.7566666666666677-2.2433333333333323 5-5 5z m-16.666666666666668-26.666666666666664c-0.9166666666666661-1.7763568394002505e-15-1.666666666666666 0.7499999999999982-1.666666666666666 1.6666666666666643v20c0 0.9166666666666679 0.75 1.6666666666666679 1.666666666666666 1.6666666666666679h16.666666666666668c0.9166666666666679 0 1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679v-20c0-0.9166666666666661-0.75-1.666666666666666-1.6666666666666679-1.666666666666666h-16.666666666666668z m15 10h-13.333333333333334c-0.46000000000000085 0-0.8333333333333339-0.37333333333333485-0.8333333333333339-0.8333333333333321s0.3733333333333331-0.8333333333333321 0.8333333333333339-0.8333333333333321h13.333333333333334c0.46000000000000085 0 0.8333333333333321 0.37333333333333485 0.8333333333333321 0.8333333333333321s-0.37333333333333485 0.8333333333333321-0.8333333333333321 0.8333333333333321z m0-5.000000000000002h-13.333333333333334c-0.46000000000000085 0-0.8333333333333339-0.3733333333333331-0.8333333333333339-0.8333333333333339s0.37333333333333485-0.8333333333333321 0.8333333333333339-0.8333333333333321h13.333333333333334c0.46000000000000085 0 0.8333333333333321 0.3733333333333331 0.8333333333333321 0.8333333333333339s-0.37333333333333485 0.8333333333333339-0.8333333333333321 0.8333333333333339z m0 10.000000000000002h-13.333333333333334c-0.46000000000000085 0-0.8333333333333339-0.37333333333333485-0.8333333333333339-0.8333333333333321s0.3733333333333331-0.8333333333333321 0.8333333333333339-0.8333333333333321h13.333333333333334c0.46000000000000085 0 0.8333333333333321 0.37333333333333485 0.8333333333333321 0.8333333333333321s-0.37333333333333485 0.8333333333333321-0.8333333333333321 0.8333333333333321z m0 5h-13.333333333333334c-0.46000000000000085 0-0.8333333333333339-0.37333333333333485-0.8333333333333339-0.8333333333333321s0.3733333333333331-0.8333333333333321 0.8333333333333339-0.8333333333333321h13.333333333333334c0.46000000000000085 0 0.8333333333333321 0.37333333333333485 0.8333333333333321 0.8333333333333321s-0.37333333333333485 0.8333333333333321-0.8333333333333321 0.8333333333333321z' })
                )
            );
        }
    }]);

    return TiDocumentText;
}(React.Component);

exports.default = TiDocumentText;
module.exports = exports['default'];