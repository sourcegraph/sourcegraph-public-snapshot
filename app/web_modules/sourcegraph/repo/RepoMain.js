// @flow weak

import React from "react";
import Helmet from "react-helmet";
import TreeSearch from "sourcegraph/tree/TreeSearch";
import Modal, {setLocationModalState} from "sourcegraph/components/Modal";
import CSSModules from "react-css-modules";
import styles from "./styles/Repo.css";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import * as BuildActions from "sourcegraph/build/BuildActions";
import "sourcegraph/build/BuildBackend";
import Dispatcher from "sourcegraph/Dispatcher";
import httpStatusCode from "sourcegraph/util/httpStatusCode";
import {trimRepo} from "sourcegraph/repo";
import context from "sourcegraph/app/context";
import {guessBranchName} from "sourcegraph/build/Build";

import Header from "sourcegraph/components/Header";

const TREE_SEARCH_MODAL_NAME = "TreeSearch";

class RepoMain extends React.Component {
	static propTypes = {
		location: React.PropTypes.object,
		repo: React.PropTypes.string,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string,
		resolvedRev: React.PropTypes.object,
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
	};

	constructor(props) {
		super(props);
		this.state = {
			treeSearchPath: "/",
			treeSearchQuery: "",
		};
		this._isMounted = false;
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._showTreeSearchModal = this._showTreeSearchModal.bind(this);
		this._dismissTreeSearchModal = this._dismissTreeSearchModal.bind(this);

		this._repoResolutionUpdated(this.props.repo, this.props.repoResolution);
		this._buildUpdated(this.props.repo, this.props.build);
	}

	state: {
		treeSearchPath: string,
		treeSearchQuery: string,
	};

	componentDidMount() {
		this._isMounted = true;
		if (global.document) {
			document.addEventListener("keydown", this._handleKeyDown);
		}
		// Whenever the user navigates to different RepoMain views, e.g.
		// navigating directories in the directory tree, viewing code
		// files, etc. we trigger a MirroredRepos.RefreshVCS operation such
		// that new changes on the remote are pulled.
		this.context.router.listenBefore(() => Dispatcher.Backends.dispatch(new RepoActions.RefreshVCS(this.props.repo)));
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

	componentWillUnmount() {
		this._isMounted = false;
		if (global.document) {
			document.removeEventListener("keydown", this._handleKeyDown);
		}
	}

	_repoResolutionUpdated(repo: string, resolution: Object) {
		// Create the repo if we don't have repoObj (the result of creating a repo) yet,
		// and this repo was just resolved to a remote repo (which must be explicitly created,
		// as we do right here).
		if (!this.props.repoObj && repo && resolution && !resolution.Error && resolution.Result.RemoteRepo) {
			// Don't create the repo if user agent is bot.
			if (context.userAgentIsBot) {
				return;
			}

			Dispatcher.Backends.dispatch(new RepoActions.WantCreateRepo(repo, resolution.Result.RemoteRepo));
		}
	}

	_buildUpdated(repo: string, build: Object) {
		// Don't trigger the build if user agent is bot.
		if (context.userAgentIsBot) {
			return;
		}

		if (build && build.Error && build.Error.response && build.Error.response.status === 404) {
			// No build exists, so create one.
			Dispatcher.Backends.dispatch(new BuildActions.CreateBuild(repo, this.props.commitID, guessBranchName(this.props.rev), null));
		}
	}

	_isMounted: boolean;
	_handleKeyDown: () => void;
	_showTreeSearchModal: () => void;
	_dismissTreeSearchModal: () => void;

	_onSelectPath(path: string) {
		this.setState({treeSearchPath: path});
	}

	_onChangeQuery(query: string) {
		this.setState({treeSearchQuery: query});
	}

	_showTreeSearchModal() {
		setLocationModalState(this.context.router, this.props.location, TREE_SEARCH_MODAL_NAME, true);
		this.setState({treeSearchPath: "/", treeSearchQuery: ""});
	}

	_dismissTreeSearchModal(loc) {
		setLocationModalState(this.context.router, this.props.location, TREE_SEARCH_MODAL_NAME, false);
	}

	_handleKeyDown(e: KeyboardEvent) {
		// Consult deepest-matched route (e.g., the "tree" subroute).
		const disableTreeSearchOverlay = this.props.routes[this.props.routes.length - 1].disableTreeSearchOverlay;

		const tag = e.target instanceof HTMLElement ? e.target.tagName : "";
		switch (e.keyCode) {
		case 84: // "t"
			if (disableTreeSearchOverlay) break;
			if (tag === "INPUT" || tag === "SELECT" || tag === "TEXTAREA") return;
			e.preventDefault();
			this._showTreeSearchModal();
			break;
		}
	}

	// canonicalURL returns the canonical URL for current page.
	canonicalURL(): string {
		// HACK: Assume that default branch name is always "master". This may not always be true,
		//       but it is for most git repos. Try this first, since it'll be accurate for most
		//       repos, and figure out the actual default branch later as a followup.
		let path = this.props.location.pathname.replace("@master/", "/");
		return `https://sourcegraph.com${path}`;
	}

	render() {
		const err = (this.props.repoResolution && this.props.repoResolution.Error) || (this.props.repoObj && this.props.repoObj.Error);
		if (err) {
			let msg;
			if (err.response && err.response.status === 401) {
				msg = `Sign in to add repositories.`;
			} else if (err.response && err.response.status === 404) {
				msg = `Repository not found.`;
			} else {
				msg = `Repository is not available.`;
			}
			return (
				<Header
					title={`${httpStatusCode(err)}`}
					subtitle={msg} />
			);
		}

		if (!this.props.repo) return null;

		if (this.props.isCloning) {
			return (
				<Header
					title="Cloning this repository"
					subtitle="Refresh this page in a minute." />
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

		let description = null;
		if (this.props.repoObj && this.props.repoObj.Description && this.props.repoObj.Description.length > 159) {
			description = this.props.repoObj.Description.substr(0, 159).concat("…");
		}
		return (
			<div>
				{description ?
					<Helmet
						title={trimRepo(this.props.repo)}
						meta={[
							{name: "description", content: description},
						]}
						link={[
							{rel: "canonical", href: this.canonicalURL()},
						]} /> :
					<Helmet
						title={trimRepo(this.props.repo)}
						link={[
							{rel: "canonical", href: this.canonicalURL()},
						]} />
				}
				{this.props.main}
				{(!this.props.route || !this.props.route.disableTreeSearchOverlay) && this.props.location.state && this.props.location.state.modal === TREE_SEARCH_MODAL_NAME &&
					<Modal onDismiss={this._dismissTreeSearchModal}>
						<div styleName="tree-search-modal">
							<TreeSearch
								repo={this.props.repo}
								rev={this.props.rev}
								commitID={this.props.commitID}
								overlay={true}
								path={this.state.treeSearchPath}
								query={this.state.treeSearchQuery}
								location={this.props.location}
								route={this.props.route}
								onChangeQuery={this._onChangeQuery.bind(this)}
								onSelectPath={this._onSelectPath.bind(this)} />
						</div>
					</Modal>
				}
			</div>
		);
	}
}

export default CSSModules(RepoMain, styles);
