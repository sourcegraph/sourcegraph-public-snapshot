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

var MdAllOut = function (_React$Component) {
    _inherits(MdAllOut, _React$Component);

    function MdAllOut() {
        _classCallCheck(this, MdAllOut);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAllOut).apply(this, arguments));
    }

    _createClass(MdAllOut, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.796666666666667 26.71666666666667q2.6566666666666663-2.655000000000001 2.6566666666666663-6.443333333333335t-2.6566666666666663-6.445-6.445-2.6566666666666663-6.443333333333333 2.6566666666666663-2.6566666666666663 6.445 2.6566666666666663 6.445 6.4449999999999985 2.6583333333333314 6.446666666666665-2.6566666666666663z m1.7966666666666669-14.683333333333335q3.4383333333333326 3.4366666666666656 3.4383333333333326 8.28t-3.4400000000000013 8.203333333333333-8.283333333333335 3.3599999999999994-8.200000000000001-3.3599999999999994-3.3616666666666664-8.203333333333333 3.3583333333333343-8.283333333333333 8.203333333333333-3.4366666666666656 8.283333333333331 3.4383333333333344z m-21.560000000000002 1.5583333333333336v-6.638333333333334h6.638333333333335z m6.638333333333335 20h-6.638333333333334v-6.640000000000001z m20-6.640000000000001v6.640000000000001h-6.638333333333335z m-6.640000000000001-20h6.638333333333335v6.640000000000001z' })
                )
            );
        }
    }]);

    return MdAllOut;
}(React.Component);

exports.default = MdAllOut;
module.exports = exports['default'];