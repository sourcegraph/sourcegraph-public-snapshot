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

var TiBook = function (_React$Component) {
    _inherits(TiBook, _React$Component);

    function TiBook() {
        _classCallCheck(this, TiBook);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiBook).apply(this, arguments));
    }

    _createClass(TiBook, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 5h-18.333333333333332c-0.44166666666666643 0-0.8666666666666671 0.17499999999999982-1.1783333333333328 0.4883333333333333l-5 5c-0.033333333333333215 0.033333333333333215-0.06666666666666643 0.06666666666666643-0.09499999999999975 0.1033333333333335-0.2333333333333334 0.27500000000000036-0.375 0.6216666666666661-0.3916666666666666 1v18.40833333333333c0 2.7566666666666677 2.243333333333334 5 5 5h15c2.173333333333332 0 4.026666666666667-1.3933333333333309 4.716666666666665-3.333333333333332h1.1166666666666671c2.336666666666666 0 4.166666666666664-2.1950000000000003 4.166666666666664-5v-16.666666666666668c0-2.756666666666667-2.2433333333333323-5-5-5z m-20 26.666666666666668c-0.9166666666666661 0-1.666666666666666-0.7466666666666661-1.666666666666666-1.6666666666666679v-16.666666666666664h3.333333333333334v18.333333333333336h-1.6666666666666679z m16.666666666666668-1.6666666666666679c0 0.9200000000000017-0.75 1.6666666666666679-1.6666666666666679 1.6666666666666679h-11.666666666666666v-18.333333333333336h11.666666666666666c0.9166666666666679 1.7763568394002505e-15 1.6666666666666679 0.7466666666666679 1.6666666666666679 1.6666666666666679v15z m5-3.333333333333332c0 1.033333333333335-0.5399999999999991 1.6666666666666679-0.8333333333333321 1.6666666666666679h-0.8333333333333357v-13.333333333333336c0-2.7566666666666677-2.2433333333333323-5-5-5h-14.31l1.666666666666666-1.666666666666666h17.643333333333334c0.9166666666666679 0 1.6666666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666v16.666666666666668z' })
                )
            );
        }
    }]);

    return TiBook;
}(React.Component);

exports.default = TiBook;
module.exports = exports['default'];