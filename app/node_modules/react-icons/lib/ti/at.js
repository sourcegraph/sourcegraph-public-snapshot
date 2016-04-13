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

var TiAt = function (_React$Component) {
    _inherits(TiAt, _React$Component);

    function TiAt() {
        _classCallCheck(this, TiAt);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiAt).apply(this, arguments));
    }

    _createClass(TiAt, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 6.666666666666667c-7.350000000000001 0-13.333333333333334 5.983333333333333-13.333333333333334 13.333333333333332s5.9833333333333325 13.333333333333336 13.333333333333334 13.333333333333336c2.693333333333335 0 5.286666666666669-0.7999999999999972 7.5-2.306666666666665 0.7583333333333329-0.5199999999999996 0.9549999999999983-1.5566666666666684 0.4366666666666674-2.3166666666666664-0.5166666666666657-0.7616666666666667-1.5533333333333346-0.9533333333333331-2.3166666666666664-0.4383333333333326-1.6566666666666663 1.1333333333333293-3.6000000000000014 1.7283333333333282-5.620000000000001 1.7283333333333282-5.516666666666666 0-10-4.483333333333334-10-10s4.483333333333334-10 10-10 10 4.483333333333334 10 10v0.8333333333333321c0 0.9200000000000017-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679s-1.6666666666666679-0.7466666666666661-1.6666666666666679-1.6666666666666679v-5c0-0.9216666666666669-0.745000000000001-1.666666666666666-1.6666666666666679-1.666666666666666-0.7333333333333343 0-1.341666666666665 0.4833333333333325-1.5633333333333326 1.1466666666666665-0.966666666666665-0.711666666666666-2.150000000000002-1.1466666666666665-3.4366666666666674-1.1466666666666665-3.216666666666665 0-5.833333333333334 2.6166666666666654-5.833333333333334 5.833333333333334s2.6166666666666654 5.833333333333336 5.833333333333334 5.833333333333336c1.7416666666666671 0 3.291666666666668-0.783333333333335 4.359999999999999-2 0.913333333333334 1.206666666666667 2.3466666666666676 2 3.9733333333333327 2 2.7566666666666677 0 5-2.2433333333333323 5-5v-0.8333333333333357c0-7.350000000000001-5.983333333333334-13.333333333333334-13.333333333333336-13.333333333333334z m0 15.833333333333332c-1.3783333333333339 0-2.5-1.1216666666666661-2.5-2.5s1.1216666666666661-2.5 2.5-2.5 2.5 1.1216666666666661 2.5 2.5-1.1216666666666661 2.5-2.5 2.5z' })
                )
            );
        }
    }]);

    return TiAt;
}(React.Component);

exports.default = TiAt;
module.exports = exports['default'];