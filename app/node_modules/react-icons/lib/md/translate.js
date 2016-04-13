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

var MdTranslate = function (_React$Component) {
    _inherits(MdTranslate, _React$Component);

    function MdTranslate() {
        _classCallCheck(this, MdTranslate);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdTranslate).apply(this, arguments));
    }

    _createClass(MdTranslate, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.483333333333334 28.36h5.391666666666666l-2.7333333333333343-7.266666666666666z m4.376666666666665-11.719999999999999l7.5 20h-3.3599999999999994l-1.875-5h-7.891666666666666l-1.8733333333333348 5h-3.3599999999999994l7.5-20h3.3599999999999994z m-9.375 8.438333333333333l-1.326666666666668 3.4383333333333326-5.158333333333331-5.156666666666666-8.360000000000001 8.283333333333331-2.3433333333333337-2.344999999999999 8.516666666666667-8.361666666666668q-3.126666666666667-3.4383333333333326-5-7.578333333333333h3.3583333333333343q1.6400000000000006 3.1250000000000036 3.828333333333333 5.548333333333336 3.5933333333333337-3.9833333333333307 5.313333333333336-8.906666666666666h-18.67166666666667v-3.3599999999999994h11.716666666666669v-3.283333333333333h3.2833333333333314v3.283333333333333h11.716666666666669v3.3599999999999994h-4.920000000000002q-2.0333333333333314 6.25-6.171666666666667 10.86l-0.07833333333333314 0.07833333333333314z' })
                )
            );
        }
    }]);

    return MdTranslate;
}(React.Component);

exports.default = MdTranslate;
module.exports = exports['default'];