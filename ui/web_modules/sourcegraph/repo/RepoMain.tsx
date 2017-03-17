import * as React from "react";
import { Route } from "react-router";

import { getRoutePattern } from "sourcegraph/app/routePatterns";
import { RouteParams, Router } from "sourcegraph/app/router";
import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { Heading, Loader } from "sourcegraph/components";
import { whitespace } from "sourcegraph/components/utils";
import { urlWithRev } from "sourcegraph/repo/routes";

import * as styles from "sourcegraph/repo/styles/Repo.css";

interface Props {
	repository: GQL.IRepository | null;
	commit: GQL.ICommitState;
	routes: Route[];
	params: RouteParams;
	location?: any;
}

export class RepoMain extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	_onKeydown = (ev: KeyboardEvent): void => {
		// Don't trigger if there's a modifier key or if the cursor is focused
		// in an input field.
		const el = ev.target as HTMLElement;
		if (!(ev.altKey || ev.ctrlKey || ev.metaKey || ev.shiftKey) &&
			typeof document !== "undefined" && el.tagName !== "INPUT" &&
			(el.tagName !== "TEXTAREA" || !isNonMonacoTextArea(el)) &&
			el.tagName !== "SELECT") {
			if (ev.keyCode === 89 /* y */ && this.props.commit.commit) {
				let url = `${urlWithRev(getRoutePattern(this.props.routes), this.props.params, this.props.commit.commit.sha1)}${window.location.hash}`;
				this.context.router.push(url);
				ev.preventDefault();
				ev.stopPropagation();
			}
		}
	}

	render(): JSX.Element {
		return (
			<div className={styles.outer_container}>
				{this.props.children}
				<EventListener target={global.document} event="keydown" callback={this._onKeydown} />
			</div>
		);
	}
}

export class CloningRefresher extends React.Component<{
	relay: any;
}, {}> {
	_refreshInterval: number | null = null;

	componentDidMount(): void {
		if (!this._refreshInterval) {
			this._refreshInterval = setInterval(this.props.relay.forceFetch, 1000);
		}
	}

	componentWillUnmount(): void {
		if (this._refreshInterval) {
			clearInterval(this._refreshInterval);
		}
	}

	render(): JSX.Element {
		return <Heading color="gray" level={4} align="center" style={{ marginTop: whitespace[5] }}>
			Cloning this repository<br />
			<Loader />
		</Heading>;
	}
}
