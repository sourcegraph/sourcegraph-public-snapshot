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

var TiFlowParallel = function (_React$Component) {
    _inherits(TiFlowParallel, _React$Component);

    function TiFlowParallel() {
        _classCallCheck(this, TiFlowParallel);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiFlowParallel).apply(this, arguments));
    }

    _createClass(TiFlowParallel, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 26.973333333333336v-13.946666666666669c1.9333333333333336-0.6933333333333334 3.3333333333333357-2.5233333333333334 3.3333333333333357-4.693333333333333 0-2.756666666666667-2.2433333333333323-5-5-5s-5 2.243333333333334-5 5c0 2.17 1.3999999999999986 4 3.333333333333332 4.691666666666666v13.950000000000001c-1.9333333333333336 0.6900000000000013-3.333333333333332 2.5216666666666683-3.333333333333332 4.691666666666666 0 2.756666666666664 2.2433333333333323 5.0000000000000036 5 5.0000000000000036s5-2.2433333333333323 5-5c0-2.1700000000000017-1.3999999999999986-4-3.333333333333332-4.693333333333335z m-1.6666666666666643-20.30666666666667c0.9200000000000017-8.881784197001252e-16 1.6666666666666679 0.7499999999999991 1.6666666666666679 1.666666666666666s-0.7466666666666661 1.666666666666666-1.6666666666666679 1.666666666666666-1.6666666666666679-0.75-1.6666666666666679-1.666666666666666 0.7466666666666661-1.666666666666667 1.6666666666666679-1.666666666666667z m0 26.666666666666668c-0.9200000000000017 0-1.6666666666666679-0.75-1.6666666666666679-1.6666666666666679s0.7466666666666661-1.6666666666666679 1.6666666666666679-1.6666666666666679 1.6666666666666679 0.75 1.6666666666666679 1.6666666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679z m-11.666666666666668-25c0-2.7566666666666686-2.243333333333334-5.000000000000002-5-5.000000000000002s-5.000000000000001 2.243333333333333-5.000000000000001 5c0 2.17 1.3999999999999995 4 3.333333333333333 4.691666666666666v13.950000000000001c-1.9333333333333336 0.6900000000000013-3.333333333333334 2.5216666666666683-3.333333333333334 4.691666666666666 0 2.756666666666664 2.243333333333334 5.0000000000000036 5 5.0000000000000036s4.999999999999998-2.2433333333333323 4.999999999999998-5c0-2.1700000000000017-1.4000000000000004-4-3.333333333333334-4.693333333333335v-13.946666666666669c1.9333333333333371-0.6933333333333334 3.3333333333333375-2.5233333333333334 3.3333333333333375-4.693333333333333z m-5-1.6666666666666687c0.9199999999999999 0 1.666666666666666 0.75 1.666666666666666 1.666666666666667s-0.7466666666666661 1.666666666666666-1.666666666666666 1.666666666666666-1.666666666666666-0.75-1.666666666666666-1.666666666666666 0.7466666666666661-1.666666666666667 1.666666666666666-1.666666666666667z m0 26.666666666666668c-0.9199999999999999 0-1.666666666666666-0.75-1.666666666666666-1.6666666666666679s0.7466666666666661-1.6666666666666679 1.666666666666666-1.6666666666666679 1.666666666666666 0.75 1.666666666666666 1.6666666666666679-0.7466666666666661 1.6666666666666679-1.666666666666666 1.6666666666666679z' })
                )
            );
        }
    }]);

    return TiFlowParallel;
}(React.Component);

exports.default = TiFlowParallel;
module.exports = exports['default'];