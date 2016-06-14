import React from "react";
import Container from "sourcegraph/Container";
import RefsContainer from "sourcegraph/def/RefsContainer";
import DefStore from "sourcegraph/def/DefStore";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import {Button} from "sourcegraph/components";
import "sourcegraph/blob/BlobBackend";
import CSSModules from "react-css-modules";
import styles from "./styles/DefInfo.css";
import {RefLocsPerPage} from "sourcegraph/def";
import "whatwg-fetch";

class RepoRefsContainer extends Container {
	static propTypes = {
		repo: React.PropTypes.string,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string,
		def: React.PropTypes.string,
		defObj: React.PropTypes.object,
		defRepos: React.PropTypes.array,
	};

	constructor(props) {
		super(props);
		this.state = {
			currPage: 1,
			nextPageLoading: false,
		};
		this._onNextPage = this._onNextPage.bind(this);
	}

	stores() {
		return [DefStore];
	}

	reconcileState(state, props) {
		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.def = props.def || null;
		state.defObj = props.defObj || null;
		state.defRepos = props.defRepos || [];
		state.refLocations = state.def ? DefStore.getRefLocations({
			repo: state.repo, commitID: state.commitID, def: state.def, repos: state.defRepos,
		}) : null;
		if (state.refLocations && state.refLocations.PagesFetched >= state.currPage) {
			state.nextPageLoading = false;
		}
	}

	onStateTransition(prevState, nextState) {
		if (nextState.currPage !== prevState.currPage || nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefLocations({
				repo: nextState.repo, commitID: nextState.commitID, def: nextState.def, repos: nextState.defRepos, page: nextState.currPage,
			}));
		}
	}

	_onNextPage() {
		let nextPage = this.state.currPage + 1;
		this.setState({currPage: nextPage, nextPageLoading: true});
		if (this.context.eventLogger) this.context.eventLogger.logEvent("RefsPaginatorClicked", {page: nextPage});
	}

	render() {
		let refLocs = this.state.refLocations;
		let fileCount = refLocs && refLocs.RepoRefs ?
			refLocs.RepoRefs.reduce((total, refs) => total + refs.Files.length, refLocs.RepoRefs[0].Files.length) : 0;

		return (
			<div>
				<div styleName="section-label">
					{refLocs && refLocs.TotalRepos &&
						`Used in ${refLocs.TotalRepos} repositor${refLocs.TotalRepos === 1 ? "y" : "ies"}`
					}
					{refLocs && !refLocs.TotalRepos && refLocs.RepoRefs &&
						`Used in ${refLocs.RepoRefs.length}+ repositories`
					}
				</div>
				<hr style={{marginTop: 0, clear: "both"}}/>
				{!refLocs && <i>Loading...</i>}
				{refLocs && !refLocs.RepoRefs && <i>No references found</i>}
				{refLocs && refLocs.RepoRefs && refLocs.RepoRefs.map((repoRefs, i) => <RefsContainer
					key={i}
					repo={this.props.repo}
					rev={this.props.rev}
					commitID={this.props.commitID}
					def={this.props.def}
					defObj={this.props.defObj}
					repoRefs={repoRefs}
					prefetch={i === 0}
					initNumSnippets={i === 0 ? 1 : 0}
					fileCollapseThreshold={5} />)}
				{/* Display the paginator if we have more files repos or repos to show. */}
				{refLocs && refLocs.RepoRefs &&
					(fileCount >= RefLocsPerPage || refLocs.TotalRepos > refLocs.RepoRefs.length || !refLocs.TotalRepos) &&
					!refLocs.StreamTerminated &&
					<div styleName="pagination">
						<Button color="blue" loading={this.state.nextPageLoading} onClick={this._onNextPage}>View More</Button>
					</div>
				}
			</div>
		);
	}
}

export default CSSModules(RepoRefsContainer, styles);
