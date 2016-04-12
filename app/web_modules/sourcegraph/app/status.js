// @flow weak

import React from "react";

// httpStatusCode returns the HTTP status code that is most appropriate
// for the given Error (or 200 for null errors);
export function httpStatusCode(err: ?Error): number {
	if (!err) return 200;
	if (err.response) return err.response.status;
	return 500;
}

export type Status = {
	reset: () => void;
	error: (err: ?Error) => void;
};

export type State = {
	// error is an error that occurred loading the current page, if any.
	error: ?Error;
};

// promises are tracked here, not in Location.state, to avoid needless component
// updates when Location.state.promises would change.
let _promises = [];
function addPromise(p: Promise): void {
	_promises.push(p);
}
function removePromise(p: Promise): void {
	let i = _promises.indexOf(p);
	if (i === -1) throw new Error(`promise is not in list`);
	_promises.splice(i, 1);
}

export function trackPromise(p: Promise): void {
	if (global.it) {
		// We're in a backend unit test, so no-op.
		return;
	}

	addPromise(p);
	p.then(() => removePromise(p));
}

// trackedPromisesCount returns the total number of tracked promises initiated
// and not yet resolved (or rejected).
//
// Only this count, not the list itself, is exported, for better encapsulation.
export function trackedPromisesCount(): number {
	return _promises.length;
}

// allTrackedPromisesResolved returns a promise that is resolved when all promises
// tracked so far are resolved. It lets server.js determine when the initial data
// loading is complete.
export function allTrackedPromisesResolved(): Promise {
	return Promise.all(_promises);
}

// withStatusContext passes a "status" context item
// to Component's children and lets them set the global HTTP response
// status code, etc. The status code is passed back to the server in
// server-side rendering, and then passed to the client. Without
// something like this, the Go HTTP handler would not know what HTTP
// status code to send to the client for server-side-rendered pages.
export function withStatusContext(Component) {
	class WithStatus extends React.Component {
		static propTypes = {
			location: React.PropTypes.object.isRequired,
		};

		static contextTypes = {
			router: React.PropTypes.object.isRequired,
		};

		static childContextTypes = {
			status: React.PropTypes.object,
		};

		getChildContext(): {status: Status} {
			return {
				status: {
					reset: (): void => {
						this._setLocationState(this.emptyState);
					},

					error: (error: ?Error): void => {
						if (!error) return;
						// Don't let a null error overwrite an existing error, so that
						// we don't clobber the error just because something else unrelated
						// succeeded.
						if (!this._getLocationState().error) this._setLocationState({error: error});
					},
				},
			};
		}

		emptyState: State = {error: null};

		_getLocationState(): State {
			return this.props.location.state ? this.props.location.state : {...this.emptyState};
		}

		_setLocationState(state: State): void {
			this.context.router.replace({
				...this.props.location,
				state: {...this._getLocationState(), ...state},
			});
		}

		render() {
			return <Component {...this.props} />;
		}
	}
	return WithStatus;
}
