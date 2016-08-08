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

var MdFindReplace = function (_React$Component) {
    _inherits(MdFindReplace, _React$Component);

    function MdFindReplace() {
        _classCallCheck(this, MdFindReplace);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFindReplace).apply(this, arguments));
    }

    _createClass(MdFindReplace, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.733333333333334 25.233333333333334l8.049999999999997 8.126666666666665-2.423333333333332 2.421666666666667-8.125-8.05q-3.1233333333333313 2.2683333333333344-6.873333333333331 2.2683333333333344-4.843333333333334 0-8.283333333333333-3.4383333333333326l-3.4366666666666683 3.4383333333333326v-10h10l-4.216666666666667 4.216666666666669q2.42 2.423333333333332 5.936666666666666 2.423333333333332 3.046666666666667 0 5.273333333333333-1.875t2.8483333333333363-4.765000000000001h3.361666666666668q-0.3133333333333326 2.8133333333333326-2.1099999999999994 5.233333333333334z m-9.371666666666666-15.233333333333334q-3.048333333333334 0-5.315000000000001 1.875t-2.8900000000000006 4.766666666666666h-3.3599999999999985q0.625-4.220000000000001 3.906666666666667-7.111666666666668t7.656666666666668-2.8900000000000006q4.766666666666666 0 8.203333333333333 3.4383333333333344l3.436666666666664-3.438333333333331v10h-10l4.219999999999999-4.216666666666667q-2.419999999999998-2.423333333333334-5.859999999999999-2.423333333333334z' })
                )
            );
        }
    }]);

    return MdFindReplace;
}(React.Component);

exports.default = MdFindReplace;
module.exports = exports['default'];