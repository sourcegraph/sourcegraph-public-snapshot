// tslint:disable

import * as React from "react";
import Helmet from "react-helmet";
import * as styles from "./styles/Repo.css";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import * as BuildActions from "sourcegraph/build/BuildActions";
import "sourcegraph/build/BuildBackend";
import Dispatcher from "sourcegraph/Dispatcher";
import httpStatusCode from "sourcegraph/util/httpStatusCode";
import {trimRepo} from "sourcegraph/repo/index";
import context from "sourcegraph/app/context";
import {guessBranchName} from "sourcegraph/build/Build";
import Header from "sourcegraph/components/Header";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

function repoPageTitle(repo: any): string {
	let title = trimRepo(repo.URI);
	if (repo.Description) {
		title += `: ${repo.Description.slice(0, 40)}${repo.Description.length > 40 ? "..." : ""}`;
	}
	return title;
}

class RepoMain extends React.Component<any, any> {
	static propTypes = {
		location: React.PropTypes.object,
		repo: React.PropTypes.string,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string,
		resolvedRev: React.PropTypes.object,
		repoNavContext: React.PropTypes.object,
		repoResolution: React.PropTypes.object,
		build: React.PropTypes.object,
		repoObj: React.PropTypes.object,
		main: React.PropTypes.element,
		isCloning: React.PropTypes.bool,
		route: React.PropTypes.object,
		routes: React.PropTypes.array,
	};

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);

		this._repoResolutionUpdated(this.props.repo, this.props.repoResolution);
		this._buildUpdated(this.props.repo, this.props.build);
	}

	componentDidMount() {
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

	componentWillReceiveProps(nextProps) {
		if (this.props.repoResolution !== nextProps.repoResolution) {
			this._repoResolutionUpdated(nextProps.repo, nextProps.repoResolution);
		}

		if (this.props.build !== nextProps.build && !this.props.build) {
			// Check for !this.props.build to avoid a loop where
			// after we create a build, this gets triggered again.
			this._buildUpdated(nextProps.repo, nextProps.build);
		}
	}

	_repoResolutionUpdated(repo: string, resolution: any) {
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

	_buildUpdated(repo: string, build: any) {
		// Don't trigger the build if user agent is bot.
		if (context.userAgentIsBot) {
			return;
		}

		if (build && build.Error && build.Error.response && build.Error.response.status === 404) {
			// No build exists, so create one.
			Dispatcher.Backends.dispatch(new BuildActions.CreateBuild(repo, this.props.commitID, guessBranchName(this.props.rev), null));
		}
	}

	render(): JSX.Element | null {
		const err = (this.props.repoResolution && this.props.repoResolution.Error) || (this.props.repoObj && this.props.repoObj.Error);
		if (err) {
			let msg;
			if (err.response && err.response.status === 401) {
				(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REPOSITORY, AnalyticsConstants.ACTION_ERROR, "ViewRepoMainError", {repo: this.props.repo, rev: this.props.rev, page_name: this.props.location.pathname, error_type: "401"});
				msg = `Sign in to add repositories.`;
			} else if (err.response && err.response.status === 404) {
				(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REPOSITORY, AnalyticsConstants.ACTION_ERROR, "ViewRepoMainError", {repo: this.props.repo, rev: this.props.rev, page_name: this.props.location.pathname, error_type: "404"});
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

		if (!this.props.repo) return null;

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

export default RepoMain;
