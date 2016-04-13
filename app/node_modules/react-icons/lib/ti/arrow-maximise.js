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

var TiArrowMaximise = function (_React$Component) {
    _inherits(TiArrowMaximise, _React$Component);

    function TiArrowMaximise() {
        _classCallCheck(this, TiArrowMaximise);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowMaximise).apply(this, arguments));
    }

    _createClass(TiArrowMaximise, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 6.666666666666667c-0.9216666666666669 0-1.6666666666666679 0.746666666666667-1.6666666666666679 1.666666666666667s0.745000000000001 1.666666666666666 1.6666666666666679 1.666666666666666h2.6433333333333344l-5.488333333333333 5.488333333333333c-0.6499999999999986 0.6499999999999986-0.6499999999999986 1.7049999999999983 0 2.3566666666666656 0.3249999999999993 0.3249999999999993 0.75 0.4883333333333333 1.1783333333333346 0.4883333333333333s0.8533333333333317-0.163333333333334 1.1783333333333346-0.4883333333333333l5.48833333333333-5.4883333333333315v2.6433333333333326c0 0.9199999999999999 0.745000000000001 1.6666666666666679 1.6666666666666679 1.6666666666666679s1.6666666666666679-0.7466666666666661 1.6666666666666679-1.666666666666666v-8.333333333333336h-8.333333333333336z m-9.511666666666667 15.488333333333333l-5.488333333333333 5.48833333333333v-2.643333333333331c0-0.9200000000000017-0.7449999999999992-1.6666666666666679-1.666666666666666-1.6666666666666679s-1.666666666666667 0.7466666666666661-1.666666666666667 1.6666666666666679v8.333333333333336h8.333333333333332c0.9199999999999999 0 1.6666666666666679-0.7466666666666697 1.6666666666666679-1.6666666666666679s-0.7449999999999992-1.6666666666666679-1.666666666666666-1.6666666666666679h-2.6433333333333344l5.488333333333335-5.486666666666665c0.6499999999999986-0.6499999999999986 0.6499999999999986-1.7049999999999983 0-2.3566666666666656s-1.7049999999999983-0.6533333333333324-2.3566666666666656 0z m-3.8216666666666654-2.155000000000001c0.9199999999999999 0 1.666666666666666-0.7466666666666661 1.666666666666666-1.6666666666666679v-4.999999999999998h5.000000000000002c0.9216666666666669 0 1.6666666666666679-0.7466666666666661 1.6666666666666679-1.666666666666666s-0.745000000000001-1.666666666666666-1.6666666666666679-1.666666666666666h-8.333333333333336v8.333333333333334c0 0.9200000000000017 0.7449999999999992 1.6666666666666679 1.666666666666666 1.6666666666666679z m16.666666666666668 0c-0.9216666666666669 0-1.6666666666666679 0.7466666666666661-1.6666666666666679 1.6666666666666679v5h-5c-0.9216666666666669 0-1.6666666666666679 0.7466666666666661-1.6666666666666679 1.6666666666666679s0.745000000000001 1.6666666666666679 1.6666666666666679 1.6666666666666679h8.333333333333332v-8.333333333333332c0-0.9200000000000017-0.745000000000001-1.6666666666666679-1.6666666666666679-1.6666666666666679z' })
                )
            );
        }
    }]);

    return TiArrowMaximise;
}(React.Component);

exports.default = TiArrowMaximise;
module.exports = exports['default'];