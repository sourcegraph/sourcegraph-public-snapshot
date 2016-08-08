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

var TiFolder = function (_React$Component) {
    _inherits(TiFolder, _React$Component);

    function TiFolder() {
        _classCallCheck(this, TiFolder);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiFolder).apply(this, arguments));
    }

    _createClass(TiFolder, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 10h-10c0-1.8399999999999999-1.4933333333333323-3.333333333333334-3.333333333333332-3.333333333333334h-6.666666666666668c-2.756666666666666 8.881784197001252e-16-5 2.243333333333334-5 5.000000000000002v16.666666666666668c0 2.7566666666666677 2.243333333333334 5 5 5h20c2.7566666666666677 0 5-2.2433333333333323 5-5v-13.333333333333336c0-2.7566666666666677-2.2433333333333323-5-5-5z m-20 0h6.666666666666668c0 1.8399999999999999 1.4933333333333323 3.333333333333334 3.333333333333332 3.333333333333334h10c0.9200000000000017 0 1.6666666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666h-23.333333333333336v-3.333333333333332c1.7763568394002505e-15-0.9199999999999999 0.7466666666666679-1.666666666666666 1.6666666666666679-1.666666666666666z m20 20h-20c-0.9199999999999999 0-1.666666666666666-0.7466666666666661-1.666666666666666-1.6666666666666679v-11.666666666666668h23.333333333333336v11.666666666666668c0 0.9200000000000017-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679z' })
                )
            );
        }
    }]);

    return TiFolder;
}(React.Component);

exports.default = TiFolder;
module.exports = exports['default'];