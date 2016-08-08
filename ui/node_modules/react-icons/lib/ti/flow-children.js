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

var TiFlowChildren = function (_React$Component) {
    _inherits(TiFlowChildren, _React$Component);

    function TiFlowChildren() {
        _classCallCheck(this, TiFlowChildren);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiFlowChildren).apply(this, arguments));
    }

    _createClass(TiFlowChildren, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.333333333333336 26.666666666666668c-2.1750000000000007 0-4.004999999999999 1.3949999999999996-4.693333333333335 3.333333333333332h-5.306666666666665c-2.7566666666666677 0-5-2.2433333333333323-5-5v-5.041666666666668c1.396666666666663 1.0583333333333336 3.1166666666666636 1.7083333333333357 5 1.7083333333333357h5.306666666666668c0.6883333333333326 1.9383333333333326 2.5199999999999996 3.333333333333332 4.693333333333332 3.333333333333332 2.7616666666666667 0 5-2.2383333333333333 5-5s-2.2383333333333333-5-5-5c-2.1750000000000007 0-4.004999999999999 1.3949999999999996-4.693333333333335 3.333333333333332h-5.306666666666665c-2.7566666666666677 0-5-2.2433333333333323-5-5v-0.30666666666666664c1.9383333333333326-0.6883333333333326 3.333333333333332-2.519999999999998 3.333333333333332-4.693333333333332 0-2.7616666666666667-2.2383333333333333-5-5-5s-5.000000000000001 2.2383333333333333-5.000000000000001 5c0 2.173333333333334 1.3950000000000005 4.005000000000001 3.333333333333333 4.693333333333333v11.973333333333333c0 4.594999999999999 3.7383333333333333 8.333333333333336 8.333333333333336 8.333333333333336h5.306666666666668c0.6883333333333326 1.9383333333333326 2.5199999999999996 3.3333333333333357 4.693333333333332 3.3333333333333357 2.7616666666666667 0 5-2.2383333333333297 5-5s-2.2383333333333333-5-5-5z m0-8.333333333333332c0.9200000000000017 0 1.6666666666666679 0.75 1.6666666666666679 1.6666666666666679s-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679-1.6666666666666679-0.75-1.6666666666666679-1.6666666666666679 0.7466666666666661-1.6666666666666679 1.6666666666666679-1.6666666666666679z m-16.666666666666668-11.666666666666668c0.9199999999999999-8.881784197001252e-16 1.666666666666666 0.7499999999999991 1.666666666666666 1.666666666666666s-0.7466666666666661 1.666666666666666-1.666666666666666 1.666666666666666-1.666666666666666-0.75-1.666666666666666-1.666666666666666 0.7466666666666661-1.666666666666667 1.666666666666666-1.666666666666667z m16.666666666666668 26.666666666666668c-0.9200000000000017 0-1.6666666666666679-0.75-1.6666666666666679-1.6666666666666679s0.7466666666666661-1.6666666666666679 1.6666666666666679-1.6666666666666679 1.6666666666666679 0.75 1.6666666666666679 1.6666666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679z' })
                )
            );
        }
    }]);

    return TiFlowChildren;
}(React.Component);

exports.default = TiFlowChildren;
module.exports = exports['default'];