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

var MdDirectionsBike = function (_React$Component) {
    _inherits(MdDirectionsBike, _React$Component);

    function MdDirectionsBike() {
        _classCallCheck(this, MdDirectionsBike);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDirectionsBike).apply(this, arguments));
    }

    _createClass(MdDirectionsBike, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 34.14000000000001q2.4216666666666633 0 4.139999999999997-1.6799999999999997t1.7166666666666686-4.100000000000001-1.7166666666666686-4.141666666666666-4.140000000000001-1.7166666666666686-4.100000000000001 1.7166666666666686-1.6833333333333336 4.141666666666666 1.6833333333333336 4.100000000000001 4.100000000000001 1.68333333333333z m0-14.14q3.5166666666666693 0 5.938333333333336 2.421666666666667t2.4216666666666598 5.9383333333333255-2.421666666666667 5.899999999999999-5.938333333333333 2.383333333333333-5.899999999999999-2.383333333333333-2.383333333333333-5.899999999999999 2.383333333333333-5.938333333333333 5.901666666666664-2.421666666666667z m-13.671666666666667-2.5l3.6733333333333356 3.828333333333333v10.313333333333333h-3.2833333333333314v-8.283333333333331l-5.388333333333334-4.686666666666667q-0.9366666666666674-0.625-0.9366666666666674-2.3433333333333337 0-1.4066666666666663 0.9383333333333326-2.3433333333333337l4.6899999999999995-4.686666666666667q0.625-0.9383333333333326 2.3433333333333337-0.9383333333333326 1.4833333333333343 0 2.6566666666666663 0.9383333333333326l3.198333333333327 3.201666666666661q2.5 2.5 5.939999999999998 2.5v3.3599999999999994q-4.923333333333332 0-8.440000000000001-3.5166666666666657l-1.3283333333333331-1.3266666666666662z m-9.606666666666667 16.64q2.421666666666667 0 4.1-1.6799999999999997t1.6833333333333336-4.100000000000001-1.6833333333333336-4.141666666666666-4.100000000000001-1.7166666666666686-4.140000000000001 1.7166666666666686-1.7166666666666668 4.141666666666666 1.7166666666666668 4.100000000000001 4.140000000000001 1.68333333333333z m0-14.14q3.5166666666666675 0 5.9 2.421666666666667t2.383333333333333 5.938333333333333-2.383333333333333 5.899999999999999-5.9 2.383333333333333-5.9383333333333335-2.383333333333333-2.4233333333333364-5.901666666666671 2.421666666666667-5.938333333333333 5.938333333333334-2.4200000000000017z m17.5-10.860000000000001q-1.3283333333333331 0-2.3433333333333337-0.9766666666666666t-1.0166666666666657-2.3049999999999997 1.0166666666666657-2.3433333333333333 2.3433333333333337-1.0166666666666666 2.3049999999999997 1.0166666666666666 0.9750000000000014 2.3416666666666663-0.9766666666666666 2.3049999999999997-2.306666666666665 0.9766666666666666z' })
                )
            );
        }
    }]);

    return MdDirectionsBike;
}(React.Component);

exports.default = MdDirectionsBike;
module.exports = exports['default'];