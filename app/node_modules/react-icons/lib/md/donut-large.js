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

var MdDonutLarge = function (_React$Component) {
    _inherits(MdDonutLarge, _React$Component);

    function MdDonutLarge() {
        _classCallCheck(this, MdDonutLarge);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDonutLarge).apply(this, arguments));
    }

    _createClass(MdDonutLarge, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 31.563333333333333q3.75-0.5466666666666669 6.600000000000001-3.3599999999999994t3.3999999999999986-6.563333333333333h5q-0.625 6.25-4.688333333333333 10.350000000000001t-10.311666666666667 4.649999999999999v-5.078333333333333z m10-13.203333333333333q-0.5466666666666669-3.75-3.3999999999999986-6.563333333333333t-6.600000000000001-3.3599999999999994v-5.078333333333333q6.25 0.6249999999999996 10.313333333333333 4.688333333333333t4.688333333333333 10.313333333333333h-5z m-13.280000000000001-9.921666666666667q-3.9066666666666663 0.625-6.953333333333333 3.9833333333333343t-3.046666666666667 7.579999999999998 3.046666666666667 7.579999999999998 6.953333333333333 3.9833333333333343v5.079999999999998q-6.328333333333333-0.625-10.663333333333334-5.390000000000001t-4.336666666666667-11.25 4.336666666666667-11.25 10.663333333333334-5.390000000000001v5.078333333333333z' })
                )
            );
        }
    }]);

    return MdDonutLarge;
}(React.Component);

exports.default = MdDonutLarge;
module.exports = exports['default'];