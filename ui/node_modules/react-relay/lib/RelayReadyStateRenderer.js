/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayReadyStateRenderer
 * 
 */

'use strict';

var _extends3 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

/**
 * @public
 *
 * RelayReadyStateRenderer synchronously renders a container and query config
 * given `readyState`. The `readyState` must be an accurate representation of
 * the data that currently resides in the supplied `environment`. If you need
 * data to be fetched in addition to rendering, please use `RelayRenderer`.
 *
 * If `readyState` is not supplied, the previously rendered `readyState` will
 * continue to be rendered (or null if there is no previous `readyState`).
 */

var RelayReadyStateRenderer = function (_React$Component) {
  (0, _inherits3['default'])(RelayReadyStateRenderer, _React$Component);

  function RelayReadyStateRenderer(props, context) {
    (0, _classCallCheck3['default'])(this, RelayReadyStateRenderer);

    var _this = (0, _possibleConstructorReturn3['default'])(this, _React$Component.call(this, props, context));

    _this.state = {
      getContainerProps: createContainerPropsFactory()
    };
    return _this;
  }

  RelayReadyStateRenderer.prototype.getChildContext = function getChildContext() {
    return {
      relay: this.props.environment,
      route: this.props.queryConfig
    };
  };

  RelayReadyStateRenderer.prototype.render = function render() {
    var children = void 0;
    var shouldUpdate = false;

    var _props = this.props;
    var readyState = _props.readyState;
    var render = _props.render;

    if (readyState) {
      if (render) {
        children = render({
          done: readyState.done,
          error: readyState.error,
          events: readyState.events,
          props: readyState.ready ? this.state.getContainerProps(this.props) : null,
          retry: this.props.retry,
          stale: readyState.stale
        });
      } else if (readyState.ready) {
        var _Container = this.props.Container;

        children = require('react').createElement(_Container, this.state.getContainerProps(this.props));
      }
      shouldUpdate = true;
    }
    if (children === undefined) {
      children = null;
      shouldUpdate = false;
    }
    return require('react').createElement(
      require('react-static-container'),
      { shouldUpdate: shouldUpdate },
      children
    );
  };

  return RelayReadyStateRenderer;
}(require('react').Component);

RelayReadyStateRenderer.childContextTypes = {
  relay: require('./RelayPropTypes').Environment,
  route: require('./RelayPropTypes').QueryConfig.isRequired
};


function createContainerPropsFactory() {
  var prevProps = void 0;
  var querySet = void 0;

  return function (nextProps) {
    if (!querySet || !prevProps || prevProps.Container !== nextProps.Container || prevProps.queryConfig !== nextProps.queryConfig) {
      querySet = require('./getRelayQueries')(nextProps.Container, nextProps.queryConfig);
    }
    var containerProps = (0, _extends3['default'])({}, nextProps.queryConfig.params, require('fbjs/lib/mapObject')(querySet, function (query) {
      return createFragmentPointerForRoot(nextProps.environment, query);
    }));
    prevProps = nextProps;
    return containerProps;
  };
}

function createFragmentPointerForRoot(environment, query) {
  return query ? require('./RelayFragmentPointer').createForRoot(environment.getStoreData().getQueuedStore(), query) : null;
}

module.exports = RelayReadyStateRenderer;