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

var FaInbox = function (_React$Component) {
    _inherits(FaInbox, _React$Component);

    function FaInbox() {
        _classCallCheck(this, FaInbox);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaInbox).apply(this, arguments));
    }

    _createClass(FaInbox, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.691428571428574 21.42857142857143h7.0528571428571425q-0.02142857142857224-0.0671428571428585-0.054285714285711606-0.17857142857142705t-0.05714285714285694-0.17857142857142705l-4.732857142857146-11.071428571428577h-15.802857142857144l-4.732857142857143 11.07142857142857q-0.02285714285714313 0.04285714285714448-0.05714285714285694 0.17857142857142705t-0.05285714285714427 0.17857142857143415h7.052857142857143l2.120000000000001 4.285714285714285h7.142857142857142z m11.451428571428572 0.6714285714285708v10.75714285714286q0 0.5799999999999983-0.42428571428571615 1.0042857142857144t-1.0042857142857144 0.42428571428570905h-31.42857142857143q-0.5799999999999992 0-1.0042857142857136-0.42428571428571615t-0.42428571428571393-1.0042857142857073v-10.75714285714286q0-1.3857142857142861 0.5571428571428574-2.747142857142858l5.314285714285715-12.321428571428571q0.22285714285714242-0.5571428571428569 0.814285714285715-0.9371428571428568t1.1714285714285708-0.3800000000000008h18.571428571428573q0.5800000000000018 0 1.1714285714285708 0.3799999999999999t0.8142857142857132 0.9371428571428568l5.314285714285717 12.321428571428573q0.5571428571428569 1.361428571428572 0.5571428571428569 2.7457142857142856z' })
                )
            );
        }
    }]);

    return FaInbox;
}(React.Component);

exports.default = FaInbox;
module.exports = exports['default'];