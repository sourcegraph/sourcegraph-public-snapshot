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

var MdControlPointDuplicate = function (_React$Component) {
    _inherits(MdControlPointDuplicate, _React$Component);

    function MdControlPointDuplicate() {
        _classCallCheck(this, MdControlPointDuplicate);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdControlPointDuplicate).apply(this, arguments));
    }

    _createClass(MdControlPointDuplicate, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 31.640000000000004q4.843333333333334 0 8.241666666666667-3.400000000000002t3.3999999999999986-8.240000000000002-3.3999999999999986-8.241666666666667-8.241666666666667-3.4000000000000004-8.241666666666667 3.4000000000000004-3.3999999999999986 8.241666666666667 3.3999999999999986 8.241666666666667 8.241666666666667 3.3999999999999986z m0-26.640000000000004q6.25 0 10.625 4.375t4.375 10.625-4.375 10.625-10.625 4.375-10.625-4.375-4.375-10.625 4.375-10.625 10.625-4.375z m-21.64 15q8.881784197001252e-16 3.4383333333333326 1.7966666666666677 6.288333333333334t4.843333333333333 4.258333333333333v3.5933333333333337q-4.375-1.5633333333333326-7.1883333333333335-5.466666666666669t-2.8116666666666665-8.673333333333332 2.811666666666667-8.671666666666667 7.188333333333333-5.466666666666667v3.5883333333333347q-3.046666666666667 1.4066666666666663-4.843333333333334 4.258333333333335t-1.796666666666666 6.291666666666664z m23.28-6.640000000000001v5h5v3.2833333333333314h-5v5h-3.2833333333333314v-5h-5v-3.2833333333333314h5v-5h3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdControlPointDuplicate;
}(React.Component);

exports.default = MdControlPointDuplicate;
module.exports = exports['default'];