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

var MdDrafts = function (_React$Component) {
    _inherits(MdDrafts, _React$Component);

    function MdDrafts() {
        _classCallCheck(this, MdDrafts);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDrafts).apply(this, arguments));
    }

    _createClass(MdDrafts, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 21.64l13.75-8.593333333333334-13.75-8.046666666666667-13.75 8.046666666666667z m16.64-8.280000000000001v16.64q0 1.3283333333333331-0.9766666666666666 2.3433333333333337t-2.3049999999999997 1.0166666666666657h-26.71666666666667q-1.3299999999999992 0-2.3066666666666658-1.0166666666666657t-0.9749999999999996-2.3433333333333337v-16.64q0-1.9533333333333314 1.5633333333333335-2.889999999999999l15.076666666666666-8.828333333333333 15.078333333333333 8.828333333333331q1.5633333333333326 0.9383333333333326 1.5633333333333326 2.8900000000000006z' })
                )
            );
        }
    }]);

    return MdDrafts;
}(React.Component);

exports.default = MdDrafts;
module.exports = exports['default'];