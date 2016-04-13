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

var MdImportantDevices = function (_React$Component) {
    _inherits(MdImportantDevices, _React$Component);

    function MdImportantDevices() {
        _classCallCheck(this, MdImportantDevices);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdImportantDevices).apply(this, arguments));
    }

    _createClass(MdImportantDevices, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm19.921666666666667 15h5.078333333333333l-4.140000000000001 2.9666666666666686 1.5633333333333326 4.844999999999999-4.063333333333333-3.046666666666667-4.140000000000001 3.0450000000000017 1.5633333333333326-4.843333333333334-4.139999999999999-2.9666666666666686h5.078333333333335l1.6400000000000006-5z m13.438333333333333-11.64q1.3283333333333331 0 2.3049999999999997 0.9383333333333335t0.9750000000000014 2.341666666666667v8.36h-3.2833333333333314v-8.356666666666666h-30v20h21.64333333333333v3.3566666666666656h-3.3599999999999994v3.3616666666666646h3.3599999999999994v3.2833333333333314h-13.36v-3.2833333333333314h3.3599999999999994v-3.3616666666666646h-11.64q-1.4066666666666658 0-2.3833333333333324-1.0133333333333319t-0.976666666666667-2.3433333333333337v-20q0-1.4066666666666663 0.9766666666666667-2.3433333333333337t2.3833333333333333-0.9383333333333335h30z m5 30v-11.716666666666669h-8.36v11.716666666666669h8.36z m0-15q0.7033333333333331 0 1.1716666666666669 0.466666666666665t0.4683333333333337 1.1733333333333356v15q0 0.7033333333333331-0.46666666666666856 1.1716666666666669t-1.173333333333332 0.46666666666666856h-8.36q-0.7033333333333331 0-1.1716666666666669-0.46666666666666856t-0.466666666666665-1.1716666666666669v-15q0-0.7033333333333331 0.466666666666665-1.1716666666666669t1.1716666666666669-0.466666666666665h8.36z' })
                )
            );
        }
    }]);

    return MdImportantDevices;
}(React.Component);

exports.default = MdImportantDevices;
module.exports = exports['default'];