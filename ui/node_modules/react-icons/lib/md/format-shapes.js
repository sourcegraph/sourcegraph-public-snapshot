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

var MdFormatShapes = function (_React$Component) {
    _inherits(MdFormatShapes, _React$Component);

    function MdFormatShapes() {
        _classCallCheck(this, MdFormatShapes);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFormatShapes).apply(this, arguments));
    }

    _createClass(MdFormatShapes, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.813333333333336 21.25h4.375l-2.188333333333336-6.4066666666666645z m5.076666666666664 2.1099999999999994h-5.859999999999999l-1.1716666666666669 3.2833333333333314h-2.7333333333333343l5.699999999999999-15h2.344999999999999l5.625 15h-2.6566666666666663z m8.75-15h3.3599999999999994v-3.3599999999999994h-3.3599999999999994v3.3599999999999994z m3.3599999999999994 26.64v-3.3599999999999994h-3.3599999999999994v3.3599999999999994h3.3599999999999994z m-6.640000000000001-3.3599999999999994v-3.2833333333333314h3.2833333333333314v-16.715000000000003h-3.2833333333333314v-3.2833333333333314h-16.71666666666667v3.283333333333333h-3.2833333333333314v16.71666666666667h3.283333333333333v3.2833333333333314h16.71666666666667z m-20 3.3599999999999994v-3.3599999999999994h-3.3599999999999994v3.3599999999999994h3.3599999999999994z m-3.3599999999999994-30v3.3599999999999994h3.3599999999999994v-3.3599999999999994h-3.3599999999999994z m33.36 6.640000000000001h-3.3599999999999994v16.716666666666665h3.3599999999999994v10.000000000000004h-10v-3.356666666666669h-16.71666666666667v3.3583333333333343h-9.999999999999998v-10h3.3566666666666674v-16.71666666666667h-3.3566666666666665v-9.999999999999998h10v3.3583333333333325h16.716666666666665v-3.36h10v10z' })
                )
            );
        }
    }]);

    return MdFormatShapes;
}(React.Component);

exports.default = MdFormatShapes;
module.exports = exports['default'];