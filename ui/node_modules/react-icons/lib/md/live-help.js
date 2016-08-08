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

var MdLiveHelp = function (_React$Component) {
    _inherits(MdLiveHelp, _React$Component);

    function MdLiveHelp() {
        _classCallCheck(this, MdLiveHelp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLiveHelp).apply(this, arguments));
    }

    _createClass(MdLiveHelp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.078333333333337 17.11q1.5633333333333326-1.5633333333333326 1.5633333333333326-3.75 0-2.7333333333333343-1.9533333333333331-4.726666666666667t-4.688333333333336-1.9916666666666663-4.688333333333334 1.9916666666666663-1.9533333333333314 4.726666666666668h3.2833333333333314q0-1.3283333333333331 1.0133333333333319-2.3433333333333337t2.3450000000000024-1.0166666666666675 2.3416666666666686 1.0166666666666657 1.0166666666666657 2.3433333333333337-1.0166666666666657 2.3433333333333337l-2.0333333333333314 2.1099999999999994q-1.9499999999999993 2.1099999999999994-1.9499999999999993 4.688333333333333v0.8616666666666681h3.280000000000001q0-2.6566666666666663 1.9533333333333331-4.766666666666666z m-3.438333333333336 12.89v-3.3599999999999994h-3.2833333333333314v3.3599999999999994h3.2833333333333314z m10-26.64q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3050000000000006v23.358333333333334q0 1.3299999999999983-1.0166666666666657 2.344999999999999t-2.3433333333333337 1.0166666666666657h-6.640000000000001l-5 5-5-5h-6.639999999999999q-1.4066666666666663 0-2.383333333333333-1.0166666666666657t-0.9766666666666683-2.344999999999999v-23.356666666666666q0-1.328333333333334 0.9766666666666666-2.3050000000000015t2.383333333333333-0.9766666666666666h23.283333333333335z' })
                )
            );
        }
    }]);

    return MdLiveHelp;
}(React.Component);

exports.default = MdLiveHelp;
module.exports = exports['default'];