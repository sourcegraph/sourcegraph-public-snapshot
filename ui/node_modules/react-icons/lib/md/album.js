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

var MdAlbum = function (_React$Component) {
    _inherits(MdAlbum, _React$Component);

    function MdAlbum() {
        _classCallCheck(this, MdAlbum);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAlbum).apply(this, arguments));
    }

    _createClass(MdAlbum, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 18.36q0.7033333333333331 0 1.1716666666666669 0.466666666666665t0.466666666666665 1.173333333333332-0.466666666666665 1.1716666666666669-1.1716666666666669 0.466666666666665-1.1716666666666669-0.466666666666665-0.466666666666665-1.1716666666666669 0.466666666666665-1.1716666666666669 1.1716666666666669-0.466666666666665z m0 9.14q3.125 0 5.313333333333333-2.1883333333333326t2.1866666666666674-5.311666666666667-2.1833333333333336-5.316666666666665-5.316666666666666-2.1833333333333353-5.313333333333334 2.1833333333333353-2.1866666666666656 5.316666666666665 2.1866666666666674 5.311666666666667 5.313333333333333 2.1883333333333326z m0-24.14q6.875 0 11.758333333333333 4.883333333333333t4.883333333333333 11.756666666666668-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z' })
                )
            );
        }
    }]);

    return MdAlbum;
}(React.Component);

exports.default = MdAlbum;
module.exports = exports['default'];