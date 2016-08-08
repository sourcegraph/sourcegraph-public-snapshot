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

var TiEject = function (_React$Component) {
    _inherits(TiEject, _React$Component);

    function TiEject() {
        _classCallCheck(this, TiEject);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiEject).apply(this, arguments));
    }

    _createClass(TiEject, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.666666666666668 29.166666666666668c-4.133333333333333 0-7.5-3.366666666666667-7.5-7.5 0-0.9216666666666669-0.7449999999999992-1.6666666666666679-1.666666666666666-1.6666666666666679s-1.666666666666666 0.745000000000001-1.666666666666666 1.6666666666666679c0 5.973333333333333 4.859999999999999 10.833333333333332 10.833333333333336 10.833333333333332s10.833333333333336-4.859999999999999 10.833333333333336-10.833333333333336-4.859999999999999-10.833333333333334-10.833333333333336-10.833333333333334c-0.9216666666666669 0-1.6666666666666679 0.7449999999999992-1.6666666666666679 1.666666666666666s0.745000000000001 1.666666666666666 1.6666666666666679 1.666666666666666c4.133333333333333 0 7.5 3.3666666666666654 7.5 7.500000000000002s-3.366666666666667 7.5-7.5 7.5z m-3.9066666666666663-22.5c0.9200000000000017-8.881784197001252e-16 1.6666666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666s-0.7466666666666661 1.666666666666666-1.6666666666666679 1.666666666666666h-5.405000000000001l9.666666666666668 9.666666666666664c0.6533333333333324 0.6499999999999986 0.6533333333333324 1.706666666666667 0.0033333333333338544 2.3566666666666656-0.31666666666666643 0.3133333333333326-0.7333333333333343 0.48666666666666814-1.1799999999999997 0.48666666666666814s-0.8666666666666671-0.173333333333332-1.1783333333333346-0.4833333333333343l-9.666666666666668-9.67333333333333v5.404999999999999c0 0.9200000000000017-0.7466666666666661 1.6666666666666679-1.666666666666666 1.6666666666666679s-1.666666666666667-0.7466666666666661-1.666666666666667-1.6666666666666679v-11.091666666666665h11.093333333333334z' })
                )
            );
        }
    }]);

    return TiEject;
}(React.Component);

exports.default = TiEject;
module.exports = exports['default'];