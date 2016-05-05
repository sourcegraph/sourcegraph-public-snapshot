// @flow weak

import React from "react";

import AuthorList from "sourcegraph/def/AuthorList";
import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import RefsContainer from "sourcegraph/def/RefsContainer";
import DefContainer from "sourcegraph/def/DefContainer";
import {Link} from "react-router";
import "sourcegraph/blob/BlobBackend";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import {urlToDef, urlToDefInfo} from "sourcegraph/def/routes";
import CSSModules from "react-css-modules";
import styles from "./styles/DefInfo.css";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import Header from "sourcegraph/components/Header";
import httpStatusCode from "sourcegraph/util/httpStatusCode";

const filesPerPage = 10;

class DefInfo extends Container {
	static contextTypes = {
		location: React.PropTypes.object,
		router: React.PropTypes.object.isRequired,
		features: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
	}

	stores() {
		return [DefStore];
	}

	componentDidMount() {
		if (super.componentDidMount) super.componentDidMount();
		// Fix a bug where navigating from a blob page here does not cause the
		// browser to scroll to the top of this page.
		if (typeof window !== "undefined") window.scrollTo(0, 0);
	}

	reconcileState(state, props) {
		state.repo = props.repo || null;
		state.rev = props.rev || null;
		state.def = props.def || null;
		state.defObj = props.defObj || null;
		state.defCommitID = props.defObj ? props.defObj.CommitID : null;
		state.refLocations = state.def ? DefStore.getRefLocations({
			repo: state.repo, rev: state.rev, def: state.def, reposOnly: false, repos: [],
			page: this._page(),
			perPage: this._perPage(),
		}) : null;
		state.authors = state.defObj ? DefStore.authors.get(state.repo, state.defObj.CommitID, state.def) : null;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefLocations({
				repo: nextState.repo, rev: nextState.rev, def: nextState.def, reposOnly: false, repos: [],
				page: this._page(),
				perPage: this._perPage(),
			}));
		}

		if (prevState.defCommitID !== nextState.defCommitID && nextState.defCommitID) {
			if (this.context.features.Authors) {
				Dispatcher.Backends.dispatch(new DefActions.WantDefAuthors(nextState.repo, nextState.defCommitID, nextState.def));
			}
		}
	}

	_page() {
		return parseInt(this.props.location.query.Page, 10) || 1;
	}

	_perPage() {
		return parseInt(this.props.location.query.PerPage, 10) || filesPerPage;
	}

	// _pageNumbers returns up to four page numbers to the left and right of
	// the given page number, clamping the result to [1, this._lastPage()].
	_pageNumbers(page) {
		// If we have less than ten pages, just show them all.
		let range = function(start, end) {
			return Array.from(new Array(end-start), (x, i) => i + start);
		};
		let pages = [];
		if (page > 1) {
			let min = page - 4;
			if (min < 1) {
				min = 1;
			}
			pages = pages.concat(range(min, page+1));
		}
		let lastPage = this._lastPage();
		if (page < lastPage) {
			let max = page + 4;
			if (max > lastPage) {
				max = lastPage;
			}
			pages = pages.concat(range(page, max+1));
		}
		return pages;
	}

	// _lastPage returns the last page number.
	_lastPage() {
		return this.state.refLocations ? Math.ceil(this.state.refLocations.TotalFiles / this._perPage()) : 0;
	}

	render() {
		let def = this.state.defObj;
		let refLocs = this.state.refLocations;
		let authors = this.state.authors;

		if (refLocs && refLocs.Error) {
			return (
				<Header
					title={`${httpStatusCode(refLocs.Error)}`}
					subtitle={`References are not available.`} />
			);
		}

		let page = this._page();
		let perPage = this._perPage();
		let lastPage = this._lastPage();
		let pages = this._pageNumbers(page);
		let pageUrl = this.state.defObj ? `${urlToDefInfo(this.state.defObj, this.state.rev)}?Page=` : "";

		let repoTimes = "";
		if (refLocs) {
			repoTimes = `${refLocs.TotalRepos} repositor${refLocs.TotalRepos > 1 ? "ies" : "y"}`;
		}

		// Under this many pages, we will hide the first/prev/next/last quick links.
		const hideFirstPrevNextLast = 10;

		return (
			<div styleName="container">
				{this.state.defObj &&
					<h1>
						<Link styleName="back-icon" to={urlToDef(this.state.defObj, this.state.rev)}>&laquo;</Link>
						<Link to={urlToDef(this.state.defObj, this.state.rev)}>
							<code styleName="def-title">{qualifiedNameAndType(this.state.defObj, {unqualifiedNameClass: styles.def})}</code>
						</Link>
					</h1>
				}
				<hr/>
				<div styleName="main">
					{authors && Object.keys(authors).length > 0 && <AuthorList authors={authors} horizontal={true} />}
					{def && def.DocHTML && <div styleName="description" dangerouslySetInnerHTML={def.DocHTML}></div>}
					{def && !def.Error && <DefContainer {...this.props} />}
					{def && !def.Error &&
						<div>
							<div styleName="section-label">
								{refLocs ? `Used in ${repoTimes}` : "Used in 0 repositories"}
								{refLocs && refLocs.TotalFiles > 1 && <span styleName="section-sub-label">(showing files {(page-1)*perPage}-{page*perPage < refLocs.TotalFiles ? page*perPage : refLocs.TotalFiles} of {refLocs.TotalFiles})</span>}
							</div>
							{!refLocs && <i>Loading...</i>}
							{refLocs && refLocs.RepoRefs.map((refRepo, i) => <RefsContainer page={page} perPage={perPage} {...this.props} key={i}
								refRepo={refRepo.Repo}
								prefetch={i === 0}
								initNumSnippets={i === 0 ? 1 : 0}
								fileCollapseThreshold={5} />)}

							{lastPage > 1 && <div styleName="pagination">
								{lastPage > hideFirstPrevNextLast && page > 1 && <Link styleName="icon" to={`${pageUrl}1`}>⇤</Link>}
								{lastPage > hideFirstPrevNextLast && page > 1 && <Link styleName="icon" to={`${pageUrl}${page-1}`}>←</Link>}
								{pages.map((n) => {
									if (page === n) {
										return <span key={n} styleName="pagination-link-disabled" to={`${pageUrl}${n}`}>{n}</span>;
									}
									return <Link key={n} styleName="pagination-link" to={`${pageUrl}${n}`}>{n}</Link>;
								})}
								{lastPage > hideFirstPrevNextLast && page < lastPage && <Link styleName="icon" to={`${pageUrl}${page+1}`}>→</Link>}
								{lastPage > hideFirstPrevNextLast && page < lastPage && <Link styleName="icon" to={`${pageUrl}${lastPage}`}>⇥</Link>}
							</div>}
						</div>
					}
				</div>
			</div>
		);
	}
}

export default CSSModules(DefInfo, styles);
