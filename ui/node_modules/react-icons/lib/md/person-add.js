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

var MdPersonAdd = function (_React$Component) {
    _inherits(MdPersonAdd, _React$Component);

    function MdPersonAdd() {
        _classCallCheck(this, MdPersonAdd);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPersonAdd).apply(this, arguments));
    }

    _createClass(MdPersonAdd, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 23.36q4.216666666666669 0 8.788333333333334 1.836666666666666t4.57 4.803333333333335v3.3616666666666646h-26.71666666666667v-3.3616666666666646q1.7763568394002505e-15-2.9666666666666686 4.56666666666667-4.803333333333335t8.791666666666664-1.836666666666666z m-15-6.719999999999999h5v3.3599999999999994h-5v5h-3.3599999999999994v-5h-5v-3.3599999999999994h5v-5h3.3599999999999994v5z m15 3.3599999999999994q-2.7333333333333343 0-4.688333333333333-1.9533333333333331t-1.9533333333333331-4.688333333333333 1.9533333333333331-4.726666666666667 4.688333333333333-1.993333333333334 4.688333333333333 1.9916666666666671 1.9533333333333331 4.725-1.9533333333333331 4.688333333333333-4.688333333333333 1.956666666666667z' })
                )
            );
        }
    }]);

    return MdPersonAdd;
}(React.Component);

exports.default = MdPersonAdd;
module.exports = exports['default'];