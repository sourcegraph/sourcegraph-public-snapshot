// @flow weak

import React from "react";

// Hack to make the status code accessible from server.js.
//
// The history object does not have a "get current state" method,
// so we can't determine the history state from the history object.

let _statusCode = null;

export function statusCode(): number { return _statusCode; }

// withHTTPResponseContext passes an "httpResponse" context item
// to Component's children and lets them set the global HTTP response
// status code. The status code is passed back to the server in
// server-side rendering, and then passed to the client. Without
// something like this, the Go HTTP handler would not know what HTTP
// status code to send to the client for server-side-rendered pages.
export function withHTTPResponseContext(Component) {
	class WithHTTPResponse extends React.Component {
		static propTypes = {
			location: React.PropTypes.object.isRequired,
		};

		static contextTypes = {
			router: React.PropTypes.object.isRequired,
		};

		static childContextTypes = {
			httpResponse: React.PropTypes.object,
		};

		getChildContext() {
			return {
				httpResponse: {
					// Don't let child components access the status code directly
					// or set it directly, since we want to ensure that only 1
					// component ever sets it. (What if one subcomponent said "200 OK"
					// but the main component said "404 Not Found"?)
					setStatusCode: (code: number) => {
						if (_statusCode && _statusCode !== code) {
							throw new Error(`Can't change HTTP response status code from ${_statusCode} to ${code}. Only one React component per request can set it.`);
						}
						_statusCode = code;

						this.context.router.replace({
							...this.props.location,
							state: {...this.props.location.state, httpStatusCode: code},
						});
					},
				},
			};
		}

		render() {
			return <Component {...this.props} />;
		}
	}
	return WithHTTPResponse;
}
