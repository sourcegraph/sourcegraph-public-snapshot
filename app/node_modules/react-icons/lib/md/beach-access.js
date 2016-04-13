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

var MdBeachAccess = function (_React$Component) {
    _inherits(MdBeachAccess, _React$Component);

    function MdBeachAccess() {
        _classCallCheck(this, MdBeachAccess);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBeachAccess).apply(this, arguments));
    }

    _createClass(MdBeachAccess, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.983333333333334 14.688333333333334q-3.9033333333333324-3.905000000000001-9.02-5.155000000000001t-9.963333333333335 0.3866666666666667q3.75-0.4666666666666668 8.241666666666667 1.4450000000000003t8.32 5.741666666666665l-9.533333333333331 9.533333333333335q-3.826666666666666-3.8299999999999983-5.74-8.321666666666665t-1.4450000000000003-8.241666666666667q-1.6400000000000006 4.843333333333334-0.39000000000000057 9.961666666666666t5.156666666666666 9.023333333333333l-4.7666666666666675 4.766666666666666q-4.92-4.923333333333332-4.92-11.916666666666668t4.921666666666667-11.911666666666667q0-0.07666666666666622 0.07833333333333314-0.07666666666666622 5.466666666666667-5.466666666666667 13.124999999999998-4.843333333333334 6.326666666666668 0.4666666666666668 10.700000000000003 4.843333333333334z m-7.186666666666667 9.608333333333333l2.421666666666667-2.421666666666667 10.704999999999998 10.783333333333331-2.423333333333332 2.3416666666666686z' })
                )
            );
        }
    }]);

    return MdBeachAccess;
}(React.Component);

exports.default = MdBeachAccess;
module.exports = exports['default'];