import * as React from "react";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import DefStore from "sourcegraph/def/DefStore";
import * as DefActions from "sourcegraph/def/DefActions";
import {Heading, List, Loader} from "sourcegraph/components";
import "sourcegraph/blob/BlobBackend";
import CSSModules from "react-css-modules";
import styles from "./styles/DefInfo.css";
import base from "sourcegraph/components/styles/_base.css";
import typography from "sourcegraph/components/styles/_typography.css";

import {Link} from "react-router";
import {refLocsPerPage} from "sourcegraph/def";
import "whatwg-fetch";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {Repository, DownPointer} from "sourcegraph/components/symbols";
import {urlToRepo} from "sourcegraph/repo/routes";
import {urlToBlob} from "sourcegraph/blob/routes";

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
		state.refLocations = props.refLocations || null;
		if (state.refLocations && state.refLocations.PagesFetched >= state.currPage) {
			state.nextPageLoading = false;
		}
	}

	onStateTransition(prevState, nextState) {
		if (nextState.currPage !== prevState.currPage || nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def || nextState.defObj !== prevState.defObj) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefLocations({
				repo: nextState.repo, commitID: nextState.defCommitID, def: nextState.def, repos: nextState.defRepos, page: nextState.currPage,
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
						`Referenced in ${refLocs.RepoRefs.length}+ repositor${refLocs.TotalRepos === 1 ? "y" : "ies"}`
					}
				</Heading>

				{!refLocs && <div className={typography.tc}> <Loader className={base.mv4} /></div>}
				{refLocs && !refLocs.RepoRefs && <i>No references found</i>}
				{refLocs && refLocs.RepoRefs && refLocs.RepoRefs.map((repoRefs, i) => <div key={i} className={base.mt4}>
					<Link to={urlToRepo(repoRefs.Repo)} className={base.mb3}>
						<Repository width={24} className={base.mr3} /> <strong>{repoRefs.Repo}</strong>
					</Link>
					<List listStyle="node" className={base.mt2} style={{marginLeft: "6px"}}>
					{repoRefs.Files && repoRefs.Files.length > 0 && repoRefs.Files.map((file, j) =>
							<li key={j} className={`${base.mv3} ${base.pt1}`} styleName="cool-mid-gray f7 node-list-item">
								{file.Count} references in <Link to={urlToBlob(repoRefs.Repo, null, file.Path)}>{file.Path}</Link>
							</li>)
					}
					</List>
				</div>)}
				{/* Display the paginator if we have more files repos or repos to show. */}
				{refLocs && refLocs.RepoRefs &&
					(fileCount >= refLocsPerPage || refLocs.TotalRepos > refLocs.RepoRefs.length || !refLocs.TotalRepos) &&
					!refLocs.StreamTerminated &&
					<a onClick={this._onNextPage} styleName="f7 link-icon">
						{this.state.nextPageLoading ?
							<strong>Loading <Loader className={base.mr2} /> </strong> :
							<strong>View more references <DownPointer styleName="icon-cool-mid-gray" width={9} className={base.ml1} /></strong>
						}
					</a>
				}
			</div>
		);
	}
}

export default CSSModules(RepoRefsContainer, styles, {allowMultiple: true});
