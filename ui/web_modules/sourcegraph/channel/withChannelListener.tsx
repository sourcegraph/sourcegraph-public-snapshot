// tslint:disable: typedef ordered-imports

import * as React from "react";

export type ChannelStatus = "connected" | "connecting" | "error" | null;
export type ChannelStatusCode = number | null;

// MAX_FAILURES is the maximum number of connection attempts to make before
// stopping.
const MAX_FAILURES = 5;

// withChannelListener wraps Component and listens on a channel for actions sent
// by the Sourcegraph API method Channel.Send. It navigates to the correct URLs
// given in the actions through the "/-/golang" route.
//
// It is used to navigate a browser in sync with an editor, for example.
//
// Assumes that the location state's "channel" property has already been
// set, by having the browser navigate to the Channel component (/-/channel?name=...).
// Once it has been set, there is no way to unset it except by reloading the
// page (this is by design; that browser tab is "controlled" by your editor).
export function withChannelListener(Component) {
	interface Props {
		location: any;
	}

	class WithChannelListener extends React.Component<Props, any> {
		static contextTypes = {
			router: React.PropTypes.object.isRequired,
		};

		state: {
			channelName: string | null;
			status: ChannelStatus;
			statusCode: ChannelStatusCode;
			failures: number;
		} = {
			channelName: null,
			status: null,
			statusCode: null,
			failures: 0,
		};

		_timeout: any;
		_unlisten: any;
		_ws: any;

		constructor(props: Props) {
			super(props);
			this._listen = this._listen.bind(this);
		}

		componentDidMount() {
			this._unlisten = (this.context as any).router.listen((loc) => {
				// The Channel component at /-/channel communicates with WithChannelListener by
				// setting the "channel" location state property. Once we see that
				// here, we add that to our state and persist it, since the "channel"
				// location state will be wiped next time we navigate.
				//
				// TODO(sqs): Figure out how aggressive to be about maintaining the channel.
				// Right now we close it when the user navigates on their own AND THEN reloads
				// the page. This seems arbitrary.
				if (loc.state && loc.state.channel && loc.state.channel !== this.state.channelName) {
					this.setState({channelName: loc.state.channel}, this._listen);
				}
			});
		}

		componentWillUnmount() {
			if (this._timeout) {
				clearTimeout(this._timeout);
				this._timeout = null;
			}
			if (this._unlisten) {
				this._unlisten();
				this._unlisten = null;
			}
			if (this._ws) {
				this._ws.close();
			}
		}

		_delayedListen(interval: number) {
			if (!this._timeout) {
				this._timeout = setTimeout(() => {
					this._timeout = null;
					this._listen();
				}, interval);
			}
		}

		_listen() {
			if (!this.state.channelName) {
				throw new Error("Unexpectedly called _listen with no channel name set in state.");
			}
			if (this._ws) {
				throw new Error("_listen called but there is an existing WebSocket conn.");
			}

			this.setState({status: "connecting"});

			const l = window.location;
			this._ws = new WebSocket(`${l.protocol === "https:" ? "wss://" : "ws://"}${l.host}/.api/channel/${encodeURIComponent(this.state.channelName)}`);
			this._ws.onopen = (ev) => {
				this.setState({status: "connected", failures: 0, statusCode: 1});
			};
			this._ws.onmessage = (ev) => {
				this._handleAction(JSON.parse(ev.data));
			};
			this._ws.onclose = (ev) => {
				this.setState({failures: this.state.failures + 1});
				if (!ev.wasClean) {
					console.error(`WebSocket closed uncleanly: ${ev.code} ${ev.reason}`);
				}
				this._ws = null;

				if (this.state.failures <= MAX_FAILURES) {
					this.setState({status: "connecting"});
					this._delayedListen(1000 + Math.pow(this.state.failures, 2) * 1000);
				} else {
					this.setState({status: "error"});
				}
			};
		}

		_handleAction(action) {
			// TODO:matt Remove special case while we switch to version numbers
			if (action && (action.VersionMajor === 0 && action.VersionMinor === 0)) {
				(this.context as any).router.push({
					pathname: `/-/channel/${this.state.channelName}-error`,
					state: Object.assign({}, this.props.location.state, {
						error: "There is an update for Sourcegraph for your editor",
						fix: "Make sure you pull the latest version of Sourcegraph for your editor.",
					}),
				});
			} else if (action && typeof action.Status !== "undefined" && action.Status === 0) {
				this.setState({statusCode: 0});
			} else if (action && action.Error && action.Fix && !action.URL) {
				(this.context as any).router.push({
					pathname: `/-/channel/${this.state.channelName}-error`,
					state: Object.assign({}, this.props.location.state, {
						error: action.Error,
						fix: action.Fix,
					}),
				});
			// TODO:matt, right documentation about how to force plugin upgrades using version numbers
			// Right now, we are checking to make sure version is >= 0.1
			} else if (action && (action.VersionMajor < 0 && action.VersionMinor < 1)) {
				(this.context as any).router.push({
					pathname: `/-/channel/${this.state.channelName}-error`,
					state: Object.assign({}, this.props.location.state, {
						error: "There is an update for Sourcegraph for your editor",
						fix: "Make sure you pull the latest version of Sourcegraph for your editor.",
					}),
				});
			} else if (action && action.Package && action.Repo && action.Status && action.Status === 1 && action.EditorType) {
				this.setState({statusCode: 1});
				let def = action.Def ? action.Def : "";
				(this.context as any).router.replace({
					pathname: "/-/golang",
					search: `?def=${def}&pkg=${action.Package}&repo=${action.Repo}&editor_type=${action.EditorType}&utm_source=sourcegraph-editor`,
					state: Object.assign({}, this.props.location.state, {
						error: null,
						fix: null,
					}),
				});
			}
		}

		render(): JSX.Element | null {
			return <Component {...this.props} {...this.state} channelStatusCode={this.state.statusCode} />;
		}
	}

	return WithChannelListener;
}
