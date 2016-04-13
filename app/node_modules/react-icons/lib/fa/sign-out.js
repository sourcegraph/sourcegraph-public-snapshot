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

var FaSignOut = function (_React$Component) {
    _inherits(FaSignOut, _React$Component);

    function FaSignOut() {
        _classCallCheck(this, FaSignOut);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaSignOut).apply(this, arguments));
    }

    _createClass(FaSignOut, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.42857142857143 32.142857142857146q0 0.09000000000000341 0.022857142857141355 0.4471428571428575t0.011428571428570677 0.5914285714285725-0.0671428571428585 0.5242857142857176-0.2228571428571442 0.4357142857142833-0.4571428571428573 0.1428571428571459h-7.142857142857144q-2.654285714285715 0-4.542857142857144-1.8857142857142861t-1.8871428571428521-4.541428571428579v-15.714285714285715q0-2.654285714285715 1.8857142857142857-4.542857142857144t4.542857142857143-1.8857142857142843h7.142857142857144q0.2914285714285718 0 0.5042857142857144 0.2142857142857144t0.21000000000000085 0.5q0 0.09142857142857164 0.024285714285714022 0.4485714285714284t0.011428571428570677 0.5914285714285716-0.0671428571428585 0.524285714285714-0.2228571428571442 0.43571428571428594-0.4571428571428573 0.1457142857142859h-7.142857142857144q-1.4714285714285715 0-2.522857142857143 1.048571428571428t-1.048571428571428 2.522857142857143v15.714285714285714q0 1.4714285714285715 1.048571428571428 2.5228571428571414t2.522857142857143 1.048571428571428h6.964285714285715l0.25714285714285623 0.022857142857141355 0.25714285714285623 0.0671428571428585 0.17857142857142705 0.12285714285714278 0.15714285714285836 0.20285714285714462 0.04285714285714448 0.29999999999999716z m20.714285714285715-12.142857142857146q0 0.5799999999999983-0.42428571428571615 1.0042857142857144l-12.142857142857142 12.142857142857139q-0.4242857142857126 0.42428571428571615-1.0042857142857144 0.42428571428571615t-1.0042857142857144-0.42428571428571615-0.42428571428571615-1.0042857142857073v-6.428571428571431h-10q-0.5800000000000001 0-1.0042857142857144-0.4242857142857126t-0.4242857142857126-1.004285714285718v-8.571428571428571q0-0.5800000000000001 0.4242857142857144-1.0042857142857144t1.0042857142857127-0.4242857142857126h10v-6.428571428571429q0-0.5800000000000001 0.4242857142857126-1.0042857142857144t1.004285714285718-0.4242857142857144 1.0042857142857144 0.4242857142857144l12.142857142857142 12.142857142857142q0.42428571428571615 0.4242857142857126 0.42428571428571615 1.0042857142857144z' })
                )
            );
        }
    }]);

    return FaSignOut;
}(React.Component);

exports.default = FaSignOut;
module.exports = exports['default'];