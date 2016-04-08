// @flow weak

import React from "react";
import TreeSearch from "sourcegraph/tree/TreeSearch";
import Modal from "sourcegraph/components/Modal";
import CSSModules from "react-css-modules";
import styles from "./styles/Repo.css";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import Dispatcher from "sourcegraph/Dispatcher";
import isMainRoute from "sourcegraph/util/isMainRoute";

import Header from "sourcegraph/components/Header";

class RepoMain extends React.Component {
	static propTypes = {
		location: React.PropTypes.object,
		repo: React.PropTypes.string,
		rev: React.PropTypes.string,
		repoObj: React.PropTypes.object,
		main: React.PropTypes.element,
		isCloning: React.PropTypes.bool,
		route: React.PropTypes.object,
		routes: React.PropTypes.array,
	};

	static contextTypes = {
		httpResponse: React.PropTypes.object,
		router: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this.state = {
			treeSearchActive: false,
			treeSearchPath: "/",
			treeSearchQuery: "",
		};
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._showTreeSearchModal = this._showTreeSearchModal.bind(this);
		this._dismissTreeSearchModal = this._dismissTreeSearchModal.bind(this);
	}

	state: {
		treeSearchActive: boolean,
		treeSearchPath: string,
		treeSearchQuery: string,
	};

	componentWillMount() {
		this._setStatusCode(this.props.repoObj, this.props.rev);
	}

	componentDidMount() {
		if (global.document) {
			document.addEventListener("keydown", this._handleKeyDown);
		}
		// Whenever the user navigates to different RepoMain views, e.g.
		// navigating directories in the directory tree, viewing code
		// files, etc. we trigger a MirroredRepos.RefreshVCS operation such
		// that new changes on the remote are pulled.
		this.context.router.listenBefore(() => Dispatcher.Backends.dispatch(new RepoActions.RefreshVCS(this.props.repo)));
		this.context.router.listenBefore(this._dismissTreeSearchModal);
	}

	componentWillReceiveProps(nextProps) {
		if (this.props.repoObj !== nextProps.repoObj || this.props.rev !== nextProps.rev) this._setStatusCode(nextProps.repoObj, nextProps.rev);
	}

	componentWillUnmount() {
		if (global.document) {
			document.removeEventListener("keydown", this._handleKeyDown);
		}
	}

	_handleKeyDown: () => void;
	_showTreeSearchModal:	() => void;
	_dismissTreeSearchModal:	() => void;

	_onSelectPath(path: string) {
		this.setState({treeSearchPath: path});
	}

	_onChangeQuery(query: string) {
		this.setState({treeSearchQuery: query});
	}

	_showTreeSearchModal() {
		this.setState({treeSearchActive: true, treeSearchPath: "/", treeSearchQuery: ""});
	}

	_dismissTreeSearchModal() {
		this.setState({treeSearchActive: false});
	}

	_handleKeyDown(e: KeyboardEvent) {
		const tag = e.target instanceof HTMLElement ? e.target.tagName : "";
		switch (e.keyCode) {
		case 84: // "t"
			if (this.props.route.disableTreeSearchOverlay) break;
			if (tag === "INPUT" || tag === "SELECT" || tag === "TEXTAREA") return;
			e.preventDefault();
			this._showTreeSearchModal();
			break;

		case 27: // ESC
			if (this.props.route.disableTreeSearchOverlay) break;
			this._dismissTreeSearchModal();
		}
	}

	_setStatusCode(repoObj, rev) {
		let statusCode = null;
		if (repoObj) {
			if (repoObj.Error) statusCode = 404;
			else if (rev) statusCode = 200;
		}
		if (isMainRoute(this.props.route, this.props.routes) || statusCode === 404) {
			// Don't clobber a subroute's error; for example, if the repo is found but
			// the def is not, we need to report 404 for the def, not 200 for the repo.
			this.context.httpResponse.setStatusCode(statusCode);
		}
	}

	render() {
		if (this.props.repoObj && this.props.repoObj.Error) {
			return (
				<Header
					title={`${this.props.repoObj.Error.Status}`}
					subtitle={`Repository "${this.props.repo}" is not available.`} />
			);
		}

		if (!this.props.repo || !this.props.rev) return null;

		if (this.props.isCloning) {
			return (
				<Header
					title="Sourcegraph is cloning this repository"
					subtitle="Refresh this page in a minute." />
			);
		}

		return (
			<div>
				{this.props.main}
				{(!this.props.route || !this.props.route.disableTreeSearchOverlay) && this.state.treeSearchActive &&
					<Modal
						shown={true}
						onDismiss={this._dismissTreeSearchModal}>
						<div styleName="tree-search-modal">
							<TreeSearch
								repo={this.props.repo}
								rev={this.props.rev}
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
