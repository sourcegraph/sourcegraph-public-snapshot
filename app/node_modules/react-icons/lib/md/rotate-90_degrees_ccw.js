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

var MdRotate90DegreesCcw = function (_React$Component) {
    _inherits(MdRotate90DegreesCcw, _React$Component);

    function MdRotate90DegreesCcw() {
        _classCallCheck(this, MdRotate90DegreesCcw);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRotate90DegreesCcw).apply(this, arguments));
    }

    _createClass(MdRotate90DegreesCcw, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.266666666666666 11.093333333333334q4.373333333333335 4.375 4.373333333333335 10.586666666666666t-4.373333333333335 10.586666666666666q-4.300000000000001 4.375-10.626666666666665 4.375-3.828333333333333 0-7.188333333333333-1.875l2.5-2.421666666666667q2.111666666666668 1.0166666666666657 4.690000000000001 1.0166666666666657 4.843333333333334 0 8.283333333333335-3.4400000000000013 3.3583333333333343-3.3599999999999994 3.3583333333333343-8.203333333333333t-3.361666666666668-8.280000000000001q-3.438333333333336-3.438333333333329-8.283333333333335-3.438333333333329v5.388333333333334l-7.030000000000001-7.033333333333333 7.033333333333331-7.105v5.390000000000001q6.326666666666668 0 10.623333333333335 4.453333333333333z m-26.096666666666664 10.39l6.093333333333332 6.094999999999999 6.091666666666667-6.093333333333334-6.088333333333333-6.091666666666665z m6.096666666666666-10.778333333333334l10.78 10.783333333333333-10.783333333333333 10.779999999999998-10.858333333333333-10.783333333333331z' })
                )
            );
        }
    }]);

    return MdRotate90DegreesCcw;
}(React.Component);

exports.default = MdRotate90DegreesCcw;
module.exports = exports['default'];