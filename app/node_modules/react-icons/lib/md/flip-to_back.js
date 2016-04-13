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

var MdFlipToBack = function (_React$Component) {
    _inherits(MdFlipToBack, _React$Component);

    function MdFlipToBack() {
        _classCallCheck(this, MdFlipToBack);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFlipToBack).apply(this, arguments));
    }

    _createClass(MdFlipToBack, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 28.36v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m0-20v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m-16.64 3.280000000000001v20h20v3.3599999999999994h-20q-1.4066666666666663 0-2.383333333333333-1.0166666666666657t-0.9766666666666666-2.34v-20h3.3599999999999994z m23.28 16.720000000000002v-3.360000000000003h3.3599999999999994q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657z m0-13.360000000000001v-3.360000000000001h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m0 6.639999999999999v-3.2833333333333314h3.3599999999999994v3.2833333333333314h-3.3599999999999994z m-16.64 6.719999999999999q-1.4066666666666663 0-2.383333333333333-1.0166666666666657t-0.9766666666666666-2.3416666666666686h3.3599999999999994v3.361666666666668z m6.640000000000001-23.36v3.3599999999999994h-3.2833333333333314v-3.3599999999999994h3.2833333333333314z m10 0q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3400000000000007h-3.361666666666668v-3.3566666666666674z m-10 20v3.3599999999999994h-3.2833333333333314v-3.3599999999999994h3.2833333333333314z m-6.640000000000001-20v3.3599999999999994h-3.3599999999999994q0-1.3283333333333331 0.9766666666666666-2.3433333333333337t2.383333333333333-1.0166666666666657z m0 13.36v3.2833333333333314h-3.3599999999999994v-3.2833333333333314h3.3599999999999994z m0-6.720000000000001v3.360000000000001h-3.3599999999999994v-3.3599999999999994h3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdFlipToBack;
}(React.Component);

exports.default = MdFlipToBack;
module.exports = exports['default'];