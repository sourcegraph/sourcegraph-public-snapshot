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

var TiCalenderOutline = function (_React$Component) {
    _inherits(TiCalenderOutline, _React$Component);

    function TiCalenderOutline() {
        _classCallCheck(this, TiCalenderOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiCalenderOutline).apply(this, arguments));
    }

    _createClass(TiCalenderOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.666666666666668 10.306666666666667v-0.30666666666666664c0-2.7616666666666667-2.2383333333333333-5-5-5s-5 2.2383333333333333-5 5h-3.333333333333332c0-2.7616666666666667-2.2383333333333333-5-5-5s-5.000000000000002 2.2383333333333333-5.000000000000002 5v0.30666666666666664c-1.9333333333333336 0.6933333333333334-3.333333333333334 2.523333333333335-3.333333333333334 4.693333333333333v15c0 2.7566666666666677 2.243333333333334 5 5 5h20c2.7566666666666677 0 5-2.2433333333333323 5-5v-15c0-2.17-1.3999999999999986-4-3.333333333333332-4.693333333333333z m-6.666666666666668-0.30666666666666664c0-0.9199999999999999 0.745000000000001-1.666666666666666 1.6666666666666679-1.666666666666666s1.6666666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666v3.333333333333334c0 0.9199999999999999-0.745000000000001 1.666666666666666-1.6666666666666679 1.666666666666666s-1.6666666666666679-0.7466666666666661-1.6666666666666679-1.666666666666666v-3.333333333333334z m-13.333333333333332 0c0-0.9199999999999999 0.7449999999999992-1.666666666666666 1.666666666666666-1.666666666666666s1.666666666666666 0.7466666666666661 1.666666666666666 1.666666666666666v3.333333333333334c0 0.9199999999999999-0.7449999999999992 1.666666666666666-1.666666666666666 1.666666666666666s-1.666666666666666-0.7466666666666661-1.666666666666666-1.666666666666666v-3.333333333333334z m20 20c0 0.9166666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679h-20c-0.9199999999999999 0-1.666666666666666-0.75-1.666666666666666-1.6666666666666679v-10h23.333333333333336v10z m0-11.666666666666668h-23.333333333333336v-3.333333333333332c1.7763568394002505e-15-0.9166666666666661 0.7466666666666679-1.666666666666666 1.6666666666666679-1.666666666666666 0 1.8399999999999999 1.493333333333334 3.333333333333334 3.333333333333334 3.333333333333334s3.333333333333334-1.493333333333334 3.333333333333334-3.333333333333334h6.666666666666668c0 1.8399999999999999 1.4933333333333323 3.333333333333334 3.333333333333332 3.333333333333334s3.333333333333332-1.493333333333334 3.333333333333332-3.333333333333334c0.9200000000000017 0 1.6666666666666679 0.75 1.6666666666666679 1.666666666666666v3.333333333333332z' })
                )
            );
        }
    }]);

    return TiCalenderOutline;
}(React.Component);

exports.default = TiCalenderOutline;
module.exports = exports['default'];