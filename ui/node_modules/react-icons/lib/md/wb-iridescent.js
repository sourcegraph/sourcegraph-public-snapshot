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

var MdWbIridescent = function (_React$Component) {
    _inherits(MdWbIridescent, _React$Component);

    function MdWbIridescent() {
        _classCallCheck(this, MdWbIridescent);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWbIridescent).apply(this, arguments));
    }

    _createClass(MdWbIridescent, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.283333333333333 33.28333333333333l-2.3450000000000006-2.423333333333332 2.966666666666667-2.9666666666666686 2.3450000000000006 2.3400000000000034z m-2.3466666666666667-25.86333333333333l2.3466666666666667-2.3416666666666677 2.966666666666667 2.9666666666666677-2.3433333333333337 2.3450000000000006z m28.125 23.516666666666666l-2.3433333333333337 2.3433333333333337-2.9666666666666686-3.0500000000000007 2.3416666666666686-2.341666666666665z m-12.423333333333332 6.483333333333334h-3.2833333333333314v-4.920000000000002h3.2833333333333314v4.921666666666667z m10.078333333333333-32.34166666666667l2.344999999999999 2.3433333333333337-2.9666666666666686 2.966666666666667-2.344999999999999-2.338333333333333z m-13.355-4.14h3.283333333333335v4.921666666666667h-3.2833333333333314v-4.921666666666667z m-10 23.2v-10h23.283333333333335v10h-23.28666666666667z' })
                )
            );
        }
    }]);

    return MdWbIridescent;
}(React.Component);

exports.default = MdWbIridescent;
module.exports = exports['default'];