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

var MdSpellcheck = function (_React$Component) {
    _inherits(MdSpellcheck, _React$Component);

    function MdSpellcheck() {
        _classCallCheck(this, MdSpellcheck);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSpellcheck).apply(this, arguments));
    }

    _createClass(MdSpellcheck, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm36.016666666666666 19.296666666666667l2.3416666666666686 2.3433333333333337-15.858333333333334 15.86-8.440000000000001-8.516666666666666 2.3433333333333337-2.3416666666666686 6.093333333333334 6.171666666666667z m-25.316666666666663-0.9366666666666674h6.875l-3.4383333333333344-9.216666666666667z m10.078333333333333 8.280000000000001l-1.9549999999999983-5h-9.370000000000005l-1.875 5h-3.5166666666666666l8.518333333333334-21.64h3.1249999999999982l8.516666666666667 21.64h-3.4400000000000013z' })
                )
            );
        }
    }]);

    return MdSpellcheck;
}(React.Component);

exports.default = MdSpellcheck;
module.exports = exports['default'];