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

var MdNaturePeople = function (_React$Component) {
    _inherits(MdNaturePeople, _React$Component);

    function MdNaturePeople() {
        _classCallCheck(this, MdNaturePeople);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdNaturePeople).apply(this, arguments));
    }

    _createClass(MdNaturePeople, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm7.5 18.36q-1.0933333333333337 0-1.7966666666666669-0.7416666666666671t-0.7033333333333331-1.756666666666666 0.7033333333333331-1.7583333333333329 1.7966666666666669-0.7416666666666671 1.7966666666666669 0.7416666666666671 0.7033333333333331 1.7583333333333329-0.7033333333333331 1.7583333333333346-1.7966666666666669 0.7433333333333323z m29.453333333333333-3.046666666666667q0 4.453333333333333-2.9666666666666686 7.733333333333334t-7.344999999999999 3.8299999999999983v6.483333333333334h5v3.2833333333333314h-26.641666666666666v-8.283333333333331h-1.6383333333333336v-6.716666666666669q0-0.7033333333333331 0.4666666666666668-1.1716666666666669t1.1733333333333338-0.466666666666665h5q0.7050000000000001 0 1.1733333333333338 0.466666666666665t0.4666666666666668 1.1716666666666669v6.716666666666669h-1.6383333333333336v5h13.361666666666668v-6.560000000000002q-4.216666666666669-0.7049999999999983-6.991666666666667-3.9466666666666654t-2.776666666666669-7.536666666666665q0-4.845000000000001 3.4383333333333344-8.283333333333333t8.283333333333335-3.438333333333334 8.241666666666667 3.438333333333334 3.3999999999999986 8.283333333333333z' })
                )
            );
        }
    }]);

    return MdNaturePeople;
}(React.Component);

exports.default = MdNaturePeople;
module.exports = exports['default'];