import * as React from "react";
import { InjectedRouter, Route } from "react-router";
import { RouteParams } from "sourcegraph/app/routeParams";
import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { Header } from "sourcegraph/components/Header";
import { PageTitle } from "sourcegraph/components/PageTitle";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import { urlWithRev } from "sourcegraph/repo/routes";
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
		if (this.props.commit.commit) {
			// prefetch on first load if repository is already cloned
			this.prefetchSymbols(this.props.repo, this.props.commit.commit);
		}
	}

	componentWillReceiveProps(nextProps: Props): void {
		this._updateRefreshInterval(nextProps.commit && nextProps.commit.cloneInProgress);
		if (nextProps.commit && nextProps.commit.commit && (!this.props.commit || !this.props.commit.commit || nextProps.commit.commit.sha1 !== this.props.commit.commit.sha1)) {
			// prefetch if the repository is being cloned or if we switch repositories after first page load
			this.prefetchSymbols(nextProps.repo, nextProps.commit.commit);
		}
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

	// prefetchSymbols best-effort prefetches symbols
	prefetchSymbols(repo: string, commit: GQL.ICommit | null): void {
		if (commit) {
			Dispatcher.Backends.dispatch(new RepoActions.WantSymbols(commit.languages, repo, commit.sha1, "", true));
		} else {
			console.error("could not fetch workspace/symbol: repository commit was null");
		}
	}

	render(): JSX.Element {
		if (!this.props.repository) {
			AnalyticsConstants.Events.ViewRepoMain_Failed.logEvent({ repo: this.props.repo, rev: this.props.rev, page_name: this.props.location.pathname, error_type: "404" });
			return (
				<div>
					<PageTitle title="Not Found" />
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

		return (
			<div className={styles.outer_container}>
				{this.props.children}
				<EventListener target={global.document.body} event="keydown" callback={this._onKeydown} />
			</div>
		);
	}
}
