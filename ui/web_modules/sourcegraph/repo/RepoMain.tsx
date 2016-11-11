import * as React from "react";
import Helmet from "react-helmet";
import {Header} from "sourcegraph/components/Header";
import {trimRepo} from "sourcegraph/repo";
import * as styles from "sourcegraph/repo/styles/Repo.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {httpStatusCode} from "sourcegraph/util/httpStatusCode";

function repoPageTitle(repo: any): string {
	let title = trimRepo(repo.URI);
	if (repo.Description) {
		title += `: ${repo.Description.slice(0, 40)}${repo.Description.length > 40 ? "..." : ""}`;
	}
	return title;
}

interface Props {
	location?: any;
	repo: string;
	rev?: string | null;
	commit: GQL.ICommitState;
	repoObj?: any;
	route?: any;
	routes: any[];
	relay: any;
}

type State = any;

export class RepoMain extends React.Component<Props, State> {
	_refreshInterval: number | null = null;

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

	render(): JSX.Element | null {
		const err = (this.props.repoObj && this.props.repoObj.Error);
		if (err) {
			let msg;
			if (err.response && err.response.status === 404) {
				AnalyticsConstants.Events.ViewRepoMain_Failed.logEvent({repo: this.props.repo, rev: this.props.rev, page_name: this.props.location.pathname, error_type: "404"});
				msg = `Repository not found.`;
			} else {
				msg = `Repository is not available.`;
			}

			return (
				<div>
				<Helmet title={"Sourcegraph - Not Found"} />
					<Header
						title={`${httpStatusCode(err)}`}
						subtitle={msg} />
				</div>
			);
		}

		if (!this.props.repo) {
			return null;
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

		const title = this.props.repoObj && !this.props.repoObj.Error ? repoPageTitle(this.props.repoObj) : null;

		return (
			<div className={styles.outer_container}>
				{/* NOTE: This should (roughly) be kept in sync with page titles in app/internal/ui. */}
				{title && <Helmet title={title} />}
				{this.props.children}
			</div>
		);
	}
}
