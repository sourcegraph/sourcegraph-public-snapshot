var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

import React from 'react';
import hoistStatics from 'hoist-non-react-statics';
import { routerShape } from './PropTypes';

function getDisplayName(WrappedComponent) {
  return WrappedComponent.displayName || WrappedComponent.name || 'Component';
}

export default function withRouter(WrappedComponent) {
  var WithRouter = React.createClass({
    displayName: 'WithRouter',

    contextTypes: { router: routerShape },
    render: function render() {
      return React.createElement(WrappedComponent, _extends({}, this.props, { router: this.context.router }));
    }
  });

  WithRouter.displayName = 'withRouter(' + getDisplayName(WrappedComponent) + ')';
  WithRouter.WrappedComponent = WrappedComponent;

  return hoistStatics(WithRouter, WrappedComponent);
}