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
import {urlToDef} from "sourcegraph/def/routes";
import CSSModules from "react-css-modules";
import styles from "./styles/DefInfo.css";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import Header from "sourcegraph/components/Header";
import httpStatusCode from "sourcegraph/util/httpStatusCode";
import Helmet from "react-helmet";
import {trimRepo} from "sourcegraph/repo";

class DefInfo extends Container {
	static contextTypes = {
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
		}) : null;
		state.authors = state.defObj ? DefStore.authors.get(state.repo, state.defObj.CommitID, state.def) : null;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefLocations({
				repo: nextState.repo, rev: nextState.rev, def: nextState.def, reposOnly: false, repos: [],
			}, {
				perPage: 50,
			}));
		}

		if (prevState.defCommitID !== nextState.defCommitID && nextState.defCommitID) {
			if (this.context.features.Authors) {
				Dispatcher.Backends.dispatch(new DefActions.WantDefAuthors(nextState.repo, nextState.defCommitID, nextState.def));
			}
		}
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

		return (
			<div styleName="container">
				{this.state.defObj && this.state.defObj.Name && <Helmet title={`${this.state.defObj.Name} · ${trimRepo(this.state.repo)}`} />}
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
							<div styleName="section-label">{`Used in ${refLocs ? `${refLocs.length} repositor${refLocs.length === 1 ? "y" : "ies"}` : ""}`}</div>
							{!refLocs && <i>Loading...</i>}
							{refLocs && refLocs.map((refRepo, i) => <RefsContainer {...this.props} key={i}
								refRepo={refRepo.Repo}
								prefetch={i === 0}
								initNumSnippets={i === 0 ? 1 : 0}
								fileCollapseThreshold={5} />)}
						</div>
					}
				</div>
			</div>
		);
	}
}

export default CSSModules(DefInfo, styles);
