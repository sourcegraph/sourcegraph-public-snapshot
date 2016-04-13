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

var MdHelp = function (_React$Component) {
    _inherits(MdHelp, _React$Component);

    function MdHelp() {
        _classCallCheck(this, MdHelp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHelp).apply(this, arguments));
    }

    _createClass(MdHelp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.078333333333337 18.75q1.5633333333333326-1.5633333333333326 1.5633333333333326-3.75 0-2.7333333333333343-1.9533333333333331-4.688333333333333t-4.688333333333336-1.953333333333335-4.688333333333334 1.953333333333335-1.9533333333333314 4.688333333333333h3.2833333333333314q0-1.3283333333333331 1.0133333333333319-2.3433333333333337t2.3450000000000024-1.0166666666666657 2.3416666666666686 1.0166666666666657 1.0166666666666657 2.3433333333333337-1.0166666666666657 2.3433333333333337l-2.0333333333333314 2.1099999999999994q-1.9499999999999993 2.1099999999999994-1.9499999999999993 4.688333333333333v0.8583333333333343h3.280000000000001q0-2.576666666666668 1.9533333333333331-4.686666666666667z m-3.4383333333333326 12.89v-3.2833333333333314h-3.283333333333335v3.2833333333333314h3.2833333333333314z m-1.6400000000000041-28.28q6.875 8.881784197001252e-16 11.758333333333333 4.883333333333335t4.883333333333333 11.756666666666666-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z' })
                )
            );
        }
    }]);

    return MdHelp;
}(React.Component);

exports.default = MdHelp;
module.exports = exports['default'];