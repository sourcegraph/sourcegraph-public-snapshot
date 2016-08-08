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

var MdMoodBad = function (_React$Component) {
    _inherits(MdMoodBad, _React$Component);

    function MdMoodBad() {
        _classCallCheck(this, MdMoodBad);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMoodBad).apply(this, arguments));
    }

    _createClass(MdMoodBad, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 23.36q2.8900000000000006 0 5.195 1.6000000000000014t3.3200000000000003 4.183333333333334h-17.031666666666666q1.0166666666666657-2.580000000000002 3.3199999999999985-4.183333333333334t5.195000000000002-1.6000000000000014z m-5.859999999999999-5q-1.0166666666666657 0-1.7583333333333329-0.7416666666666671t-0.7433333333333341-1.7566666666666677 0.7416666666666671-1.7583333333333329 1.756666666666666-0.7400000000000002 1.7583333333333329 0.7416666666666671 0.7399999999999984 1.7599999999999998-0.7416666666666671 1.7583333333333329-1.7599999999999998 0.7433333333333323z m11.719999999999999 0q-1.0166666666666657 0-1.7583333333333329-0.7416666666666671t-0.7399999999999984-1.7566666666666677 0.7416666666666671-1.7583333333333329 1.7600000000000016-0.7400000000000002 1.7583333333333329 0.7416666666666671 0.7433333333333323 1.7599999999999998-0.7416666666666671 1.7583333333333329-1.7566666666666677 0.7433333333333323z m-5.859999999999999 15q5.466666666666669 0 9.413333333333334-3.9450000000000003t3.9450000000000003-9.415-3.9450000000000003-9.411666666666667-9.413333333333334-3.9449999999999994-9.413333333333332 3.9449999999999994-3.945000000000001 9.411666666666667 3.9449999999999994 9.416666666666668 9.413333333333334 3.9416666666666664z m0-30q6.875 0 11.758333333333333 4.883333333333333t4.883333333333333 11.756666666666668-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z' })
                )
            );
        }
    }]);

    return MdMoodBad;
}(React.Component);

exports.default = MdMoodBad;
module.exports = exports['default'];