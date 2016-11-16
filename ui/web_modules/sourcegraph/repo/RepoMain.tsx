import * as React from "react";
import Helmet from "react-helmet";
import {InjectedRouter, Route} from "react-router";
import {RouteParams} from "sourcegraph/app/routeParams";
import {EventListener, isNonMonacoTextArea} from "sourcegraph/Component";
import {Header} from "sourcegraph/components/Header";
import {trimRepo} from "sourcegraph/repo";
import {urlWithRev} from "sourcegraph/repo/routes";
import * as styles from "sourcegraph/repo/styles/Repo.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props {
	repo: string;
	rev?: string | null;
	repository: GQL.IRepository | null;
	commit: GQL.ICommitState;
	routes: Route[];
	params: RouteParams;
	location?: any;
	relay: any;
}

export class RepoMain extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: InjectedRouter };

	_refreshInterval: number | null = null;

	constructor(props: Props) {
		super(props);
		this._onKeydown = this._onKeydown.bind(this);
	}

	componentDidMount(): void {
		this._updateRefreshInterval(this.props.commit && this.props.commit.cloneInProgress);
	}

	componentWillReceiveProps(nextProps: Props): void {
		this._updateRefreshInterval(nextProps.commit && nextProps.commit.cloneInProgress);
	}

	componentWillUnmount(): void {
		if (this._refreshInterval) {
			clearInterval(this._refreshInterval);
		}
	}

	_updateRefreshInterval(cloneInProgress: boolean): void {
		if (cloneInProgress) {
			if (!this._refreshInterval) {
				this._refreshInterval = setInterval(() => {
					this.props.relay.forceFetch();
				}, 1000);
			}
		} else {
			if (this._refreshInterval) {
				clearInterval(this._refreshInterval);
				this._refreshInterval = null;
			}
		}
	}

	_onKeydown(ev: KeyboardEvent): void {
		// Don't trigger if there's a modifier key or if the cursor is focused
		// in an input field.
		const el = ev.target as HTMLElement;
		if (!(ev.altKey || ev.ctrlKey || ev.metaKey || ev.shiftKey) &&
			typeof document !== "undefined" && el.tagName !== "INPUT" &&
			(el.tagName !== "TEXTAREA" || !isNonMonacoTextArea(el)) &&
			el.tagName !== "SELECT") {
			if (ev.keyCode === 89 /* y */ && this.props.commit.commit) {
				let url = `${urlWithRev(this.props.routes, this.props.params, this.props.commit.commit.sha1)}${window.location.hash}`;
				this.context.router.push(url);
				ev.preventDefault();
				ev.stopPropagation();
			}
		}
	}

	render(): JSX.Element {
		if (!this.props.repository) {
			AnalyticsConstants.Events.ViewRepoMain_Failed.logEvent({repo: this.props.repo, rev: this.props.rev, page_name: this.props.location.pathname, error_type: "404"});
			return (
				<div>
					<Helmet title="Not Found" />
					<Header
						title="404"
						subtitle="Repository not found." />
				</div>
			);
		}

		if (this.props.commit.cloneInProgress) {
			return (
				<Header title="Cloning this repository" loading={true} />
			);
		}

		if (!this.props.commit.commit) {
			return (
				<Header title="404"
					subtitle="Revision is not available." />
			);
		}

		let title = trimRepo(this.props.repo);
		let description = this.props.repository.description;
		if (description) {
			title += `: ${description.slice(0, 40)}${description.length > 40 ? "..." : ""}`;
		}

		return (
			<div className={styles.outer_container}>
				{/* NOTE: This should (roughly) be kept in sync with page titles in app/internal/ui. */}
				{title && <Helmet title={title} />}
				{this.props.children}
				<EventListener target={global.document.body} event="keydown" callback={this._onKeydown} />
			</div>
		);
	}
}
