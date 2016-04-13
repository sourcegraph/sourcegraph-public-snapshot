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

var FaFile = function (_React$Component) {
    _inherits(FaFile, _React$Component);

    function FaFile() {
        _classCallCheck(this, FaFile);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFile).apply(this, arguments));
    }

    _createClass(FaFile, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.714285714285715 11.428571428571429v-10.535714285714286q0.4914285714285711 0.31428571428571495 0.8028571428571425 0.6257142857142863l9.107142857142861 9.107142857142858q0.3142857142857167 0.31428571428571495 0.6257142857142881 0.8028571428571425h-10.535714285714292z m-2.8571428571428577 0.7142857142857135q0 0.8928571428571423 0.6257142857142846 1.5171428571428578t1.5171428571428578 0.6257142857142863h12.142857142857146v23.571428571428577q0 0.8928571428571459-0.6257142857142881 1.5171428571428578t-1.5171428571428578 0.625714285714281h-30q-0.8928571428571432 0-1.5171428571428573-0.6257142857142881t-0.6257142857142854-1.5171428571428507v-35.714285714285715q0-0.8928571428571459 0.6257142857142859-1.51714285714286t1.517142857142857-0.6257142857142859h17.857142857142858v12.142857142857142z' })
                )
            );
        }
    }]);

    return FaFile;
}(React.Component);

exports.default = FaFile;
module.exports = exports['default'];