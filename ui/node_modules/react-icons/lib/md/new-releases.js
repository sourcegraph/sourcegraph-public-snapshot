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

var MdNewReleases = function (_React$Component) {
    _inherits(MdNewReleases, _React$Component);

    function MdNewReleases() {
        _classCallCheck(this, MdNewReleases);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdNewReleases).apply(this, arguments));
    }

    _createClass(MdNewReleases, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 21.64v-10h-3.2833333333333314v10h3.2833333333333314z m0 6.719999999999999v-3.3599999999999994h-3.2833333333333314v3.3599999999999994h3.2833333333333314z m16.72-8.36l-4.063333333333333 4.609999999999999 0.5466666666666669 6.171666666666667-6.016666666666666 1.326666666666668-3.123333333333335 5.313333333333333-5.703333333333333-2.421666666666667-5.703333333333333 2.423333333333332-3.125-5.233333333333334-6.016666666666667-1.408333333333335 0.5499999999999998-6.173333333333336-4.068333333333333-4.608333333333327 4.066666666666666-4.688333333333333-0.5499999999999998-6.093333333333334 6.016666666666667-1.326666666666667 3.125-5.3133333333333335 5.705 2.421666666666667 5.703333333333333-2.4233333333333333 3.125 5.313333333333334 6.016666666666666 1.3283333333333331-0.5499999999999972 6.173333333333334z' })
                )
            );
        }
    }]);

    return MdNewReleases;
}(React.Component);

exports.default = MdNewReleases;
module.exports = exports['default'];