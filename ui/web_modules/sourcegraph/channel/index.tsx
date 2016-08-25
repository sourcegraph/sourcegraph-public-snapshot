// tslint:disable: typedef ordered-imports

import * as React from "react";
import {PlainRoute} from "react-router";
import {Header} from "sourcegraph/components/Header";
import * as styles from "sourcegraph/channel/styles/index.css";

interface Props {
	location: any;
	params: any;
}

type State = any;

class Channel extends React.Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
		features: React.PropTypes.object.isRequired,
	};

	_timeout: any;

	constructor(props: Props) {
		super(props);
		this.state = {takingAWhile: false};
	}

	componentDidMount(): void {
		this._timeout = setTimeout(() => {
			this.setState({takingAWhile: true});
		}, 3000);

		if (this.props.params.channel) {
			// Add channel to location state. withChannelListener will
			// notice this and respond to actions sent on that channel.
			(this.context as any).router.replace(Object.assign({}, this.props.location, {
				state: Object.assign({}, this.props.location.state, {
					channel: this.props.params.channel,
				}),
			}));
		}
	}

	componentWillUnmount(): void {
		if (this._timeout) {
			clearTimeout(this._timeout);
		}
	}

	render(): JSX.Element | null {
		if (this.props.location.state && this.props.location.state.error && this.props.location.state.fix) {
			return (
				<div className={styles.error}>
				<Header title={this.props.location.state.error}
					subtitle={this.props.location.state.fix} />
				<a href="https://github.com/sourcegraph/sourcegraph-sublime" className={styles.readme}>Sourcegraph Sublime README</a>
				</div>
			);
		}

		return (
			<Header title="We've turned this into a desktop application! This will require an upgrade by August 31, 2016." subtitle="Email beta@sourcegraph.com to migrate to the new version."/>
		);
	}
}

export const routes: PlainRoute[] = [
	{
		path: "-/channel/:channel",
		components: {
			main: Channel,
		},
	},

	// Backcompat redirect for old /-/live/:channel URLs.
	//
	// Remove this soon as the old URL was used only for limited testing.
	{
		path: "-/live/:channel",
		onEnter: (nextState, replace) => {
			replace(Object.assign({}, nextState.location, {pathname: nextState.location.pathname.replace("/-/live/", "/-/channel/")}));
		},
	},
];
