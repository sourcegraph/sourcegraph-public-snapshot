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

var MdCopyright = function (_React$Component) {
    _inherits(MdCopyright, _React$Component);

    function MdCopyright() {
        _classCallCheck(this, MdCopyright);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCopyright).apply(this, arguments));
    }

    _createClass(MdCopyright, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 33.36q5.466666666666669 0 9.413333333333334-3.9450000000000003t3.9450000000000003-9.415-3.9450000000000003-9.411666666666667-9.413333333333334-3.9449999999999994-9.413333333333332 3.9449999999999994-3.945000000000001 9.411666666666667 3.9449999999999994 9.416666666666668 9.413333333333334 3.9416666666666664z m0-30q6.875 0 11.758333333333333 4.883333333333333t4.883333333333333 11.756666666666668-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z m-0.23333333333333428 11.875q-3.126666666666665 0-3.126666666666665 4.533333333333335v0.466666666666665q0 4.533333333333335 3.125 4.533333333333335 1.1700000000000017 0 1.9499999999999993-0.663333333333334t0.783333333333335-1.6799999999999997h2.9666666666666686q0 1.9533333333333331-1.7166666666666686 3.4383333333333326-1.6400000000000006 1.4066666666666663-3.9833333333333343 1.4066666666666663-3.126666666666665 0-4.766666666666666-1.8733333333333348t-1.6400000000000006-5.156666666666666v-0.466666666666665q0-3.2049999999999983 1.5633333333333326-5 1.875-2.1116666666666664 4.843333333333334-2.1116666666666664 2.578333333333333 0 4.063333333333333 1.4833333333333325 1.6400000000000006 1.6416666666666657 1.6400000000000006 3.830000000000002h-2.9666666666666686q0-0.5466666666666669-0.23666666666666814-1.0166666666666657-0.39000000000000057-0.7800000000000011-0.5466666666666669-0.9366666666666674-0.783333333333335-0.7799999999999994-1.9533333333333331-0.7799999999999994z' })
                )
            );
        }
    }]);

    return MdCopyright;
}(React.Component);

exports.default = MdCopyright;
module.exports = exports['default'];