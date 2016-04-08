// @flow weak

import React from "react";
import TreeSearch from "sourcegraph/tree/TreeSearch";
import CSSModules from "react-css-modules";
import styles from "./styles/Repo.css";
import Component from "sourcegraph/Component";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import Dispatcher from "sourcegraph/Dispatcher";

import Header from "sourcegraph/components/Header";

class RepoMain extends Component {
	static propTypes = {
		location: React.PropTypes.object,
		repo: React.PropTypes.string,
		rev: React.PropTypes.string,
		repoObj: React.PropTypes.object,
		main: React.PropTypes.element,
		isCloning: React.PropTypes.bool,
		route: React.PropTypes.object,
	};

	constructor(props) {
		super(props);
		this._hasMounted = false;
	}

	componentDidMount() {
		this._hasMounted = true;
		Dispatcher.Backends.dispatch(new RepoActions.RefreshVCS(this.state.repo));
	}

	reconcileState(state, props) {
		super.reconcileState();
		Object.assign(state, props);
	}

	onStateTransition(prevState, nextState) {
		if (this._hasMounted && prevState.location.pathname !== nextState.location.pathname) {
			// Whenever the user navigates to different RepoMain views, e.g.
			// navigating directories in the directory tree, viewing code
			// files, etc. we trigger a MirroredRepos.RefreshVCS operation such
			// that new changes on the remote are pulled.
			Dispatcher.Backends.dispatch(new RepoActions.RefreshVCS(this.state.repo));
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
				{this.props.route.disableTreeSearchOverlay ? null : <TreeSearch repo={this.props.repo} rev={this.props.rev} path="/" overlay={true} location={this.props.location} route={this.props.route} />}
			</div>
		);
	}
}

export default CSSModules(RepoMain, styles);
