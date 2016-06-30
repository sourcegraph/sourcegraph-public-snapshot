import React from "react";
import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import {Button, Heading} from "sourcegraph/components";
import "sourcegraph/blob/BlobBackend";
import CSSModules from "react-css-modules";
import styles from "./styles/DefInfo.css";
import base from "sourcegraph/components/styles/_base.css";
import {RefLocsPerPage} from "sourcegraph/def";
import "whatwg-fetch";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {Repository} from "sourcegraph/components/symbols";

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
		if (this.context.eventLogger) this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DEF_INFO, AnalyticsConstants.ACTION_CLICK, "RefsPaginatorClicked", {next_page: nextPage, repo: this.props.repo, def: this.props.def});
	}

	render() {
		let refLocs = this.state.refLocations;
		let fileCount = refLocs && refLocs.RepoRefs ?
			refLocs.RepoRefs.reduce((total, refs) => total + refs.Files.length, refLocs.RepoRefs[0].Files.length) : 0;

		return (
			<div>
				<Heading level="7" styleName="cool-mid-gray" className={base.mb4}>
					{refLocs && refLocs.TotalRepos &&
						`Referenced in ${refLocs.TotalRepos} repositor${refLocs.TotalRepos === 1 ? "y" : "ies"}`
					}
					{refLocs && !refLocs.TotalRepos && refLocs.RepoRefs &&
						`Referenced in ${refLocs.RepoRefs.length}+ repositories`
					}
				</Heading>

				{!refLocs && <i>Loading...</i>}
				{refLocs && !refLocs.RepoRefs && <i>No references found</i>}
				{refLocs && refLocs.RepoRefs && refLocs.RepoRefs.map((repoRefs, i) => <div key={i} className={base.mb4}>
						<Repository width={24} styleName="v-mid" className={base.mr3} /> {repoRefs.Repo}
						{repoRefs.Files && repoRefs.Files.length > 0 &&
							repoRefs.Files.map((file, j) => <div key={j}>{file.Count} references in {file.Path}</div>)
						}
				</div>)}
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
