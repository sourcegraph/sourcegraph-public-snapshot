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

var MdLightbulbOutline = function (_React$Component) {
    _inherits(MdLightbulbOutline, _React$Component);

    function MdLightbulbOutline() {
        _classCallCheck(this, MdLightbulbOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLightbulbOutline).apply(this, arguments));
    }

    _createClass(MdLightbulbOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm24.766666666666666 21.796666666666667q3.591666666666665-2.5 3.591666666666665-6.796666666666667 0-3.4383333333333344-2.461666666666666-5.9t-5.900000000000002-2.46-5.896666666666661 2.46-2.461666666666668 5.9q0 4.296666666666667 3.5933333333333337 6.796666666666667l1.4049999999999994 1.0166666666666657v3.826666666666668h6.716666666666669v-3.828333333333333z m-4.766666666666666-18.436666666666667q4.843333333333334 8.881784197001252e-16 8.241666666666667 3.4000000000000012t3.3999999999999986 8.239999999999998q0 6.093333333333334-5 9.533333333333331v3.826666666666668q0 0.7033333333333331-0.46999999999999886 1.1716666666666669t-1.1716666666666669 0.4683333333333337h-10q-0.7033333333333331 0-1.1716666666666669-0.466666666666665t-0.4666666666666668-1.173333333333332v-3.826666666666668q-5-3.4400000000000013-5-9.533333333333333 0-4.843333333333334 3.3966666666666665-8.241666666666667t8.241666666666667-3.400000000000001z m-5 31.64v-1.6400000000000006h10v1.6400000000000006q0 0.7033333333333331-0.466666666666665 1.1716666666666669t-1.173333333333332 0.46666666666666856h-6.716666666666669q-0.7050000000000001 0-1.1733333333333338-0.46666666666666856t-0.466666666666665-1.1716666666666669z' })
                )
            );
        }
    }]);

    return MdLightbulbOutline;
}(React.Component);

exports.default = MdLightbulbOutline;
module.exports = exports['default'];