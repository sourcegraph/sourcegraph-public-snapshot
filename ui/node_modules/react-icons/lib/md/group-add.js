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

var MdGroupAdd = function (_React$Component) {
    _inherits(MdGroupAdd, _React$Component);

    function MdGroupAdd() {
        _classCallCheck(this, MdGroupAdd);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdGroupAdd).apply(this, arguments));
    }

    _createClass(MdGroupAdd, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 21.64q3.125 0 6.563333333333333 1.3666666666666671t3.4383333333333326 3.633333333333333v3.3599999999999994h-20v-3.3599999999999994q0-2.2666666666666657 3.4383333333333344-3.633333333333333t6.563333333333334-1.3666666666666671z m11.094999999999999 0.3133333333333326q2.8916666666666657 0.466666666666665 5.079999999999998 1.6799999999999997t2.1850000000000023 3.008333333333333v3.3583333333333343h-5v-3.3583333333333343q0-2.6566666666666663-2.2633333333333354-4.688333333333333z m-11.093333333333334-3.5933333333333337q-2.0333333333333314 0-3.5166666666666657-1.4833333333333343t-1.4833333333333343-3.5166666666666657 1.4833333333333343-3.5166666666666657 3.5166666666666657-1.4833333333333343 3.5166666666666657 1.4833333333333343 1.4833333333333343 3.5166666666666657-1.4833333333333343 3.5166666666666657-3.5166666666666657 1.4833333333333343z m8.36 0q-0.783333333333335 0-1.4833333333333343-0.23333333333333428 1.4833333333333343-2.111666666666668 1.4833333333333343-4.766666666666666t-1.4833333333333343-4.766666666666666q0.6999999999999993-0.2333333333333325 1.4833333333333343-0.2333333333333325 2.0333333333333314 0 3.5166666666666657 1.4833333333333325t1.4833333333333343 3.5166666666666657-1.4833333333333343 3.5166666666666657-3.5166666666666657 1.4833333333333343z m-16.641666666666666-1.7199999999999989v3.3599999999999994h-4.999999999999998v5h-3.360000000000001v-5h-5v-3.3599999999999994h5v-5h3.3599999999999994v5h5z' })
                )
            );
        }
    }]);

    return MdGroupAdd;
}(React.Component);

exports.default = MdGroupAdd;
module.exports = exports['default'];