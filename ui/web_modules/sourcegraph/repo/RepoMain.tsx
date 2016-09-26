import * as React from "react";
import Helmet from "react-helmet";

import * as Dispatcher from "sourcegraph/Dispatcher";

import {trimRepo} from "sourcegraph/repo";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import * as styles from "sourcegraph/repo/styles/Repo.css";

import {context} from "sourcegraph/app/context";

import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import {Header} from "sourcegraph/components/Header";

import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";
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
	rev?: string;
	commitID?: string;
	resolvedRev?: any;
	repoNavContext?: any;
	repoResolution?: any;
	repoObj?: any;
	main?: JSX.Element;
	isCloning?: boolean;
	route?: any;
	routes: any[];
}

type State = any;

export class RepoMain extends React.Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props: Props) {
		super(props);

		this._repoResolutionUpdated(this.props.repo, this.props.repoResolution);
	}

	componentDidMount(): void {
		// Whenever the user navigates to different RepoMain views, e.g.
		// navigating directories in the directory tree, viewing code
		// files, etc. we trigger a MirroredRepos.RefreshVCS operation such
		// that new changes on the remote are pulled.
		(this.context as any).router.listenBefore((loc) => {
			// Don't incur the overhead of triggering this when only the internal state changes.
			if (loc.pathname !== this.props.location.pathname) {
				Dispatcher.Backends.dispatch(new RepoActions.RefreshVCS(this.props.repo));
			}
		});
	}

	componentWillReceiveProps(nextProps: Props): void {
		if (this.props.repoResolution !== nextProps.repoResolution) {
			this._repoResolutionUpdated(nextProps.repo, nextProps.repoResolution);
		}
	}

	_repoResolutionUpdated(repo: string, resolution: any): void {
		// Create the repo if we don't have repoObj (the result of creating a repo) yet,
		// and this repo was just resolved to a remote repo (which must be explicitly created,
		// as we do right here).
		if (!this.props.repoObj && repo && resolution && !resolution.Error && !resolution.Repo && resolution.RemoteRepo) {
			// Don't create the repo if user agent is bot.
			if (context.userAgentIsBot) {
				return;
			}

			Dispatcher.Backends.dispatch(new RepoActions.WantCreateRepo(repo, resolution.RemoteRepo));
		}
	}

	render(): JSX.Element | null {
		const err = (this.props.repoResolution && this.props.repoResolution.Error) || (this.props.repoObj && this.props.repoObj.Error);
		if (err) {
			let msg;
			let showGitHubCTA = false;
			if (err.response && err.response.status === 401) {
				EventLogger.logNonInteractionEventForCategory(AnalyticsConstants.CATEGORY_REPOSITORY, AnalyticsConstants.ACTION_ERROR, "ViewRepoMainError", {repo: this.props.repo, rev: this.props.rev, page_name: this.props.location.pathname, error_type: "401"});
				msg = context.user ? `Connect GitHub to add repositories` : `Sign in to add repositories.`;
				showGitHubCTA = Boolean(context.user && !context.hasPrivateGitHubToken());
			} else if (err.response && err.response.status === 404) {
				EventLogger.logNonInteractionEventForCategory(AnalyticsConstants.CATEGORY_REPOSITORY, AnalyticsConstants.ACTION_ERROR, "ViewRepoMainError", {repo: this.props.repo, rev: this.props.rev, page_name: this.props.location.pathname, error_type: "404"});
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
				{this.props.main}
			</div>
		);
	}
}
