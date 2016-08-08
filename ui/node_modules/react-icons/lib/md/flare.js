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

var MdFlare = function (_React$Component) {
    _inherits(MdFlare, _React$Component);

    function MdFlare() {
        _classCallCheck(this, MdFlare);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFlare).apply(this, arguments));
    }

    _createClass(MdFlare, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.36 38.36v-10h3.2833333333333314v10h-3.2833333333333314z m-8.983333333333333-10.078333333333333l3.591666666666667-3.594999999999999 2.3450000000000006 2.3433333333333337-3.5933333333333337 3.591666666666665z m15.311666666666666-1.25l2.3433333333333337-2.344999999999999 3.591666666666665 3.5933333333333337-2.3433333333333337 2.3416666666666686z m-4.688333333333333-12.031666666666666q2.0333333333333314 0 3.5166666666666657 1.4833333333333343t1.4833333333333343 3.5166666666666657-1.4833333333333343 3.5166666666666657-3.5166666666666657 1.4833333333333343-3.5166666666666657-1.4833333333333343-1.4833333333333343-3.5166666666666657 1.4833333333333343-3.5166666666666657 3.5166666666666657-1.4833333333333343z m8.36 3.3599999999999994h10v3.2833333333333314h-10v-3.2833333333333314z m2.2666666666666657-6.641666666666666l-3.594999999999999 3.5950000000000006-2.344999999999999-2.3433333333333337 3.5933333333333337-3.591666666666667z m-8.98833333333333-10.076666666666668v10h-3.2833333333333314v-10h3.2833333333333314z m-6.325000000000001 11.325000000000001l-2.3433333333333337 2.3466666666666676-3.5933333333333337-3.5966666666666676 2.3433333333333337-2.341666666666667z m-3.673333333333334 5.393333333333333v3.2833333333333314h-10v-3.2833333333333314h10z' })
                )
            );
        }
    }]);

    return MdFlare;
}(React.Component);

exports.default = MdFlare;
module.exports = exports['default'];