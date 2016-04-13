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

var TiShoppingBag = function (_React$Component) {
    _inherits(TiShoppingBag, _React$Component);

    function TiShoppingBag() {
        _classCallCheck(this, TiShoppingBag);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiShoppingBag).apply(this, arguments));
    }

    _createClass(TiShoppingBag, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.333333333333336 6.666666666666667h-16.666666666666668c-2.7566666666666677 0-5.000000000000001 2.243333333333333-5.000000000000001 5.000000000000001v18.333333333333336c0 2.756666666666664 2.243333333333333 4.9999999999999964 5.000000000000001 4.9999999999999964h16.666666666666668c2.7566666666666677 0 5-2.2433333333333323 5-5v-18.333333333333332c0-2.7566666666666677-2.2433333333333323-5-5-5z m1.6666666666666679 23.333333333333336c0 0.9166666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679h-16.666666666666668c-0.9199999999999999 0-1.666666666666666-0.75-1.666666666666666-1.6666666666666679v-12.133333333333336c0.49333333333333407 0.28999999999999915 1.0583333333333336 0.466666666666665 1.666666666666666 0.466666666666665h2.5c0 3.216666666666665 2.616666666666667 5.833333333333332 5.833333333333332 5.833333333333332s5.833333333333336-2.616666666666667 5.833333333333336-5.833333333333336h2.5c0.6083333333333343 0 1.173333333333332-0.17666666666666586 1.6666666666666679-0.466666666666665v12.133333333333336z m-14.166666666666668-11.666666666666668h8.333333333333336c0 2.296666666666667-1.8666666666666671 4.166666666666668-4.166666666666668 4.166666666666668s-4.16666666666667-1.870000000000001-4.16666666666667-4.166666666666668z m14.166666666666664-3.3333333333333357c0 0.9166666666666661-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679h-16.666666666666664c-0.9199999999999999 0-1.666666666666666-0.75-1.666666666666666-1.666666666666666v-3.333333333333334c0-0.9166666666666661 0.7466666666666661-1.666666666666666 1.666666666666666-1.666666666666666h16.666666666666668c0.9200000000000017 0 1.6666666666666679 0.75 1.6666666666666679 1.666666666666666v3.333333333333334z' })
                )
            );
        }
    }]);

    return TiShoppingBag;
}(React.Component);

exports.default = TiShoppingBag;
module.exports = exports['default'];