/**
 * Copyright (c) 2013-present, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule RelayRenderer
 * 
 */

'use strict';

var _extends3 = _interopRequireDefault(require('babel-runtime/helpers/extends'));

var _classCallCheck3 = _interopRequireDefault(require('babel-runtime/helpers/classCallCheck'));

var _possibleConstructorReturn3 = _interopRequireDefault(require('babel-runtime/helpers/possibleConstructorReturn'));

var _inherits3 = _interopRequireDefault(require('babel-runtime/helpers/inherits'));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var PropTypes = require('react').PropTypes;

var INACTIVE_READY_STATE = {
  aborted: false,
  done: false,
  error: null,
  events: [],
  ready: false,
  stale: false
};

/**
 * @public
 *
 * RelayRenderer renders a container and query config after fulfilling its data
 * dependencies. Precise rendering behavior is configured via the `render` prop
 * which takes a callback.
 *
 * The container created using `Relay.createContainer` must be supplied via the
 * `Container` prop, and the query configuration that conforms to the shape of a
 * `RelayQueryConfig` must be supplied via the `queryConfig` prop.
 *
 * === Render Callback ===
 *
 * The `render` callback is called with an object with the following properties:
 *
 *   props: ?{[propName: string]: mixed}
 *     If present, sufficient data is ready to render the container. This object
 *     must be spread into the container using the spread attribute operator. If
 *     absent, there is insufficient data to render the container.
 *
 *   done: boolean
 *     Whether all data dependencies have been fulfilled. If `props` is present
 *     but `done` is false, then sufficient data is ready to render, but some
 *     data dependencies have not yet been fulfilled.
 *
 *   error: ?Error
 *     If present, an error occurred while fulfilling data dependencies. If
 *     `props` and `error` are both present, then sufficient data is ready to
 *     render, but an error occurred while fulfilling deferred dependencies.
 *
 *   retry: ?Function
 *     A function that can be called to re-attempt to fulfill data dependencies.
 *     This property is only present if an `error` has occurred.
 *
 *   stale: boolean
 *     When `forceFetch` is enabled, a request is always made to fetch updated
 *     data. However, if all data dependencies can be immediately fulfilled, the
 *     `props` property will be present. In this case, `stale` will be true.
 *
 * The `render` callback can return `undefined` to continue rendering the last
 * view rendered (e.g. when transitioning from one `queryConfig` to another).
 *
 * If a `render` callback is not supplied, the default behavior is to render the
 * container if data is available, the existing view if one exists, or nothing.
 *
 * === Refs ===
 *
 * References to elements rendered by the `render` callback can be obtained by
 * using the React `ref` prop. For example:
 *
 *   <FooComponent {...props} ref={handleFooRef} />
 *
 *   function handleFooRef(component) {
 *     // Invoked when `<FooComponent>` is mounted or unmounted. When mounted,
 *     // `component` will be the component. When unmounted, `component` will
 *     // be null.
 *   }
 *
 */

var RelayRenderer = function (_React$Component) {
  (0, _inherits3['default'])(RelayRenderer, _React$Component);

  function RelayRenderer(props, context) {
    (0, _classCallCheck3['default'])(this, RelayRenderer);

    var _this = (0, _possibleConstructorReturn3['default'])(this, _React$Component.call(this, props, context));

    var garbageCollector = _this.props.environment.getStoreData().getGarbageCollector();
    _this.gcHold = garbageCollector && garbageCollector.acquireHold();
    _this.mounted = true;
    _this.pendingRequest = null;
    _this.state = {
      active: false,
      readyState: null,
      retry: _this._retry.bind(_this)
    };
    return _this;
  }

  RelayRenderer.prototype.componentDidMount = function componentDidMount() {
    this._runQueries(this.props);
  };

  /**
   * @private
   */


  RelayRenderer.prototype._runQueries = function _runQueries(_ref) {
    var _this2 = this;

    var Container = _ref.Container;
    var forceFetch = _ref.forceFetch;
    var onForceFetch = _ref.onForceFetch;
    var onPrimeCache = _ref.onPrimeCache;
    var queryConfig = _ref.queryConfig;
    var environment = _ref.environment;

    var onReadyStateChange = function onReadyStateChange(readyState) {
      if (!_this2.mounted) {
        _this2._handleReadyStateChange((0, _extends3['default'])({}, readyState, { mounted: false }));
        return;
      }
      if (request !== _this2.lastRequest) {
        // Ignore (abort) ready state if we have a new pending request.
        return;
      }
      if (readyState.aborted || readyState.done || readyState.error) {
        _this2.pendingRequest = null;
      }
      _this2.setState({
        active: true,
        readyState: (0, _extends3['default'])({}, readyState, {
          mounted: true
        })
      });
    };

    if (this.pendingRequest) {
      this.pendingRequest.abort();
    }

    var querySet = require('./getRelayQueries')(Container, queryConfig);
    var request = this.pendingRequest = forceFetch ? onForceFetch ? onForceFetch(querySet, onReadyStateChange) : environment.forceFetch(querySet, onReadyStateChange) : onPrimeCache ? onPrimeCache(querySet, onReadyStateChange) : environment.primeCache(querySet, onReadyStateChange);
    this.lastRequest = request;
  };

  /**
   * @private
   */


  RelayRenderer.prototype._retry = function _retry() {
    var readyState = this.state.readyState;

    if (readyState && readyState.error) {
      this._runQueries(this.props);
      this.setState({ readyState: null });
    }
  };

  RelayRenderer.prototype.componentWillReceiveProps = function componentWillReceiveProps(nextProps) {
    if (nextProps.Container !== this.props.Container || nextProps.environment !== this.props.environment || nextProps.queryConfig !== this.props.queryConfig || nextProps.forceFetch && !this.props.forceFetch) {
      if (nextProps.environment !== this.props.environment) {
        if (this.gcHold) {
          this.gcHold.release();
        }
        var garbageCollector = nextProps.environment.getStoreData().getGarbageCollector();
        this.gcHold = garbageCollector && garbageCollector.acquireHold();
      }
      this._runQueries(nextProps);
      this.setState({ readyState: null });
    }
  };

  RelayRenderer.prototype.componentDidUpdate = function componentDidUpdate(prevProps, prevState) {
    // `prevState` should exist; the truthy check is for Flow soundness.
    var readyState = this.state.readyState;

    if (readyState) {
      if (!prevState || readyState !== prevState.readyState) {
        this._handleReadyStateChange(readyState);
      }
    }
  };

  /**
   * @private
   */


  RelayRenderer.prototype._handleReadyStateChange = function _handleReadyStateChange(readyState) {
    var onReadyStateChange = this.props.onReadyStateChange;

    if (onReadyStateChange) {
      onReadyStateChange(readyState);
    }
  };

  RelayRenderer.prototype.componentWillUnmount = function componentWillUnmount() {
    if (this.pendingRequest) {
      this.pendingRequest.abort();
    }
    if (this.gcHold) {
      this.gcHold.release();
    }
    this.gcHold = null;
    this.mounted = false;
  };

  RelayRenderer.prototype.render = function render() {
    var readyState = this.state.active ? this.state.readyState : INACTIVE_READY_STATE;

    return require('react').createElement(require('./RelayReadyStateRenderer'), {
      Container: this.props.Container,
      environment: this.props.environment,
      queryConfig: this.props.queryConfig,
      readyState: readyState,
      render: this.props.render,
      retry: this.state.retry
    });
  };

  return RelayRenderer;
}(require('react').Component);

RelayRenderer.propTypes = {
  Container: require('./RelayPropTypes').Container,
  forceFetch: PropTypes.bool,
  onReadyStateChange: PropTypes.func,
  queryConfig: require('./RelayPropTypes').QueryConfig.isRequired,
  environment: require('./RelayPropTypes').Environment,
  render: PropTypes.func
};

module.exports = RelayRenderer;