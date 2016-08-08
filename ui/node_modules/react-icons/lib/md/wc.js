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

var MdWc = function (_React$Component) {
    _inherits(MdWc, _React$Component);

    function MdWc() {
        _classCallCheck(this, MdWc);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWc).apply(this, arguments));
    }

    _createClass(MdWc, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.5 10q-1.4066666666666663 0-2.383333333333333-0.9766666666666666t-0.9766666666666666-2.383333333333333 0.9766666666666666-2.3433333333333337 2.383333333333333-0.9383333333333335 2.383333333333333 0.9383333333333335 0.9766666666666666 2.3433333333333337-0.9766666666666666 2.383333333333333-2.383333333333333 0.9766666666666666z m-15 0q-1.4066666666666663 0-2.383333333333333-0.9766666666666666t-0.9766666666666666-2.383333333333333 0.9766666666666666-2.3433333333333337 2.383333333333333-0.9383333333333335 2.383333333333333 0.9383333333333335 0.9766666666666666 2.3433333333333337-0.9766666666666666 2.383333333333333-2.383333333333333 0.9766666666666666z m17.5 26.640000000000008h-5v-10h-5l4.216666666666669-12.656666666666668q0.8616666666666681-2.3433333333333337 3.2049999999999983-2.3433333333333337h0.1566666666666663q2.3433333333333337 0 3.203333333333333 2.3433333333333337l4.218333333333334 12.656666666666661h-5v10z m-20.86 0v-12.5h-2.5v-9.140000000000008q0-1.3283333333333331 1.0166666666666666-2.3433333333333337t2.3416666666666677-1.0166666666666657h5q1.326666666666668 0 2.3416666666666686 1.0166666666666657t1.0183333333333309 2.3433333333333337v9.14h-2.5v12.5h-6.721666666666668z' })
                )
            );
        }
    }]);

    return MdWc;
}(React.Component);

exports.default = MdWc;
module.exports = exports['default'];