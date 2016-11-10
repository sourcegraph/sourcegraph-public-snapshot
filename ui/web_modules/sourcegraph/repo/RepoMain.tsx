import * as React from "react";
import Helmet from "react-helmet";
import {context} from "sourcegraph/app/context";
import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import {Header} from "sourcegraph/components/Header";
import {trimRepo} from "sourcegraph/repo";
import * as styles from "sourcegraph/repo/styles/Repo.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {httpStatusCode} from "sourcegraph/util/httpStatusCode";
import {privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";

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
	resolvedRev?: any;
	repoObj?: any;
	isCloning?: boolean;
	route?: any;
	routes: any[];
}

type State = any;

export class RepoMain extends React.Component<Props, State> {
	render(): JSX.Element | null {
		const err = (this.props.repoObj && this.props.repoObj.Error);
		if (err) {
			let msg;
			let showGitHubCTA = false;
			if (err.response && err.response.status === 401) {
				AnalyticsConstants.Events.ViewRepoMain_Failed.logEvent({repo: this.props.repo, rev: this.props.rev, page_name: this.props.location.pathname, error_type: "401"});
				msg = context.user ? `Connect GitHub to add repositories` : `Sign in to add repositories.`;
				showGitHubCTA = Boolean(context.user && !context.hasPrivateGitHubToken());
			} else if (err.response && err.response.status === 404) {
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
					{showGitHubCTA &&
						<div style={{textAlign: "center"}}>
							<GitHubAuthButton scopes={privateGitHubOAuthScopes} returnTo={this.props.location}>Add private repositories</GitHubAuthButton>
						</div>
					}
				</div>
			);
		}

		if (!this.props.repo) {
			return null;
		}

		if (this.props.isCloning) {
			return (
				<Header title="Cloning this repository" loading={true} />
			);
		}

		if (this.props.resolvedRev && this.props.resolvedRev.Error) {
			const err2 = this.props.resolvedRev.Error;
			const msg = `Revision is not available.`;
			return (
				<Header title={`${httpStatusCode(err2)}`}
					subtitle={msg} />
			);
		}

		// Determine if the repo route is the main route (not one of its
		// children like DefInfo, for example).
		const mainRoute = this.props.routes[this.props.routes.length - 1];
		const isMainRoute = mainRoute === this.props.route.indexRoute || mainRoute === this.props.route.indexRoute;
		const title = this.props.repoObj && !this.props.repoObj.Error ? repoPageTitle(this.props.repoObj) : null;

		return (
			<div className={styles.outer_container}>
				{/* NOTE: This should (roughly) be kept in sync with page titles in app/internal/ui. */}
				{isMainRoute && title && <Helmet title={title} />}
				{this.props.children}
			</div>
		);
	}
}
