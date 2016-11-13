import * as React from "react";
import Helmet from "react-helmet";
import {Header} from "sourcegraph/components/Header";
import {trimRepo} from "sourcegraph/repo";
import * as styles from "sourcegraph/repo/styles/Repo.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props {
	repo: string;
	rev?: string | null;
	repository: GQL.IRepository | null;
	commit: GQL.ICommitState;
	location?: any;
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
		if (!this.props.repository) {
			AnalyticsConstants.Events.ViewRepoMain_Failed.logEvent({repo: this.props.repo, rev: this.props.rev, page_name: this.props.location.pathname, error_type: "404"});
			return (
				<div>
					<Helmet title="Sourcegraph - Not Found" />
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
			</div>
		);
	}
}
