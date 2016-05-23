// @flow weak

import React from "react";
import Helmet from "react-helmet";
import AuthorList from "sourcegraph/def/AuthorList";
import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import RefsContainer from "sourcegraph/def/RefsContainer";
import DefContainer from "sourcegraph/def/DefContainer";
import {RefLocsPerPage} from "sourcegraph/def";
import {Button} from "sourcegraph/components";
import {Link} from "react-router";
import "sourcegraph/blob/BlobBackend";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import {urlToDef} from "sourcegraph/def/routes";
import CSSModules from "react-css-modules";
import styles from "./styles/DefInfo.css";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import Header from "sourcegraph/components/Header";
import httpStatusCode from "sourcegraph/util/httpStatusCode";
import {trimRepo} from "sourcegraph/repo";
import {defTitle, defTitleOK} from "sourcegraph/def/Formatter";

class DefInfo extends Container {
	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		features: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	static propTypes = {
		repo: React.PropTypes.string,
		def: React.PropTypes.string.isRequired,
		commitID: React.PropTypes.string,
		rev: React.PropTypes.string,
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
		state.authors = state.defObj ? DefStore.authors.get(state.repo, state.defObj.CommitID, state.def) : null;

		state.refLocations = state.def ? DefStore.getRefLocations({
			repo: state.repo, commitID: state.commitID, def: state.def, repos: [],
		}) : null;
		if (state.refLocations && state.refLocations.PagesFetched >= state.currPage) {
			state.nextPageLoading = false;
		}
	}

	onStateTransition(prevState, nextState) {
		if (nextState.currPage !== prevState.currPage || nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefLocations({
				repo: nextState.repo, commitID: nextState.commitID, def: nextState.def, repos: [], page: nextState.currPage,
			}));
		}

		if (prevState.defCommitID !== nextState.defCommitID && nextState.defCommitID) {
			if (this.context.features.Authors) {
				Dispatcher.Backends.dispatch(new DefActions.WantDefAuthors(nextState.repo, nextState.defCommitID, nextState.def));
			}
		}
	}

	_onNextPage() {
		let nextPage = this.state.currPage + 1;
		this.setState({currPage: nextPage, nextPageLoading: true});
		this.context.eventLogger.logEvent("RefsPaginatorClicked", {page: nextPage});
	}

	render() {
		let def = this.state.defObj;
		let refLocs = this.state.refLocations;
		let authors = this.state.authors;
		let fileCount = refLocs && refLocs.RepoRefs ?
			refLocs.RepoRefs.reduce((total, refs) => total + refs.Files.length, refLocs.RepoRefs[0].Files.length) : 0;

		if (refLocs && refLocs.Error) {
			return (
				<Header
					title={`${httpStatusCode(refLocs.Error)}`}
					subtitle={`References are not available.`} />
			);
		}

		let title = trimRepo(this.state.repo);
		let description_title = trimRepo(this.state.repo);
		if (defTitleOK(def)) {
			title = `${defTitle(def)} · ${trimRepo(this.state.repo)}`;
			description_title = `${defTitle(def)} in ${trimRepo(this.state.repo)}`;
		}
		let description = `Code and usage examples for ${description_title}.`;
		if (def && def.Docs && def.Docs.length && def.Docs[0].Data) {
			description = description.concat(" ").concat(def.Docs[0].Data);
		}
		if (description.length > 159) {
			description = description.substring(0, 159).concat("…");
		}
		return (

			<div styleName="container">
				{description ?
					<Helmet
						title={title}
						meta={[
							{name: "description", content: description},
						]} /> :
					<Helmet title={title} />
				}
				{def &&
					<h1 styleName="def-header">
						<Link title="View definition in code" styleName="back-icon" to={urlToDef(def, this.state.rev)}>&laquo;</Link>
						&nbsp;
						<Link to={urlToDef(def, this.state.rev)}>
							<code styleName="def-title">{qualifiedNameAndType(def, {unqualifiedNameClass: styles.def})}</code>
						</Link>
					</h1>
				}
				<hr/>
				<div>
					{authors && Object.keys(authors).length > 0 && <AuthorList authors={authors} horizontal={true} />}
					{def && def.DocHTML && <div styleName="description" dangerouslySetInnerHTML={def.DocHTML}></div>}
					{/* TODO DocHTML will not be set if the this def was loaded via the
						serveDefs endpoint instead of the serveDef endpoint. In this case
						we'll fallback to displaying plain text. We should be able to
						sanitize/render DocHTML on the front-end to make this consistent.
					*/}
					{def && !def.DocHTML && def.Docs && def.Docs.length &&
						<div styleName="description">{def.Docs[0].Data}</div>
					}
					{def && !def.Error && <DefContainer {...this.props} />}
					{def && !def.Error &&
						<div>
							{!refLocs && <i>Loading...</i>}
							{refLocs && refLocs.TotalRepos &&
								<div styleName="section-label">
									Used in {refLocs.TotalRepos} repositor{refLocs.TotalRepos === 1 ? "y" : "ies"}
								</div>
							}
							{refLocs && !refLocs.TotalRepos && refLocs.RepoRefs &&
								<div styleName="section-label">
									Used in {refLocs.RepoRefs.length}+ repositories
								</div>
							}
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
						</div>
					}
				</div>
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

export default CSSModules(DefInfo, styles);
