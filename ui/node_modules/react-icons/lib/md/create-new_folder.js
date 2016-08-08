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

var MdCreateNewFolder = function (_React$Component) {
    _inherits(MdCreateNewFolder, _React$Component);

    function MdCreateNewFolder() {
        _classCallCheck(this, MdCreateNewFolder);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCreateNewFolder).apply(this, arguments));
    }

    _createClass(MdCreateNewFolder, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 23.36v-3.3599999999999994h-5v-5h-3.283333333333335v5h-5v3.3599999999999994h5v5h3.2833333333333314v-5h5z m1.7199999999999953-13.36q1.4066666666666663 0 2.3433333333333337 0.9766666666666666t0.9383333333333326 2.383333333333333v16.64q0 1.4066666666666663-0.9383333333333326 2.383333333333333t-2.3433333333333337 0.9766666666666666h-26.716666666666665q-1.408333333333334 0-2.3450000000000006-0.9766666666666666t-0.94-2.383333333333333v-20q0-1.4066666666666663 0.938333333333333-2.383333333333333t2.341666666666667-0.9766666666666666h10.000000000000002l3.3616666666666646 3.3599999999999994h13.361666666666665z' })
                )
            );
        }
    }]);

    return MdCreateNewFolder;
}(React.Component);

exports.default = MdCreateNewFolder;
module.exports = exports['default'];