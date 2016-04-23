// @flow weak

import React from "react";

import AuthorList from "sourcegraph/def/AuthorList";
import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import {Link} from "react-router";
import "sourcegraph/blob/BlobBackend";
import Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import {urlToDef} from "sourcegraph/def/routes";
import CSSModules from "react-css-modules";
import styles from "./styles/DefInfo.css";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import {urlToDefRefs} from "sourcegraph/def/routes";
import Header from "sourcegraph/components/Header";
import {httpStatusCode} from "sourcegraph/app/status";
import ctx from "sourcegraph/app/context";

class DefInfo extends Container {
	static contextTypes = {
		status: React.PropTypes.object,
		router: React.PropTypes.object.isRequired,
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
		state.refLocations = state.def ? DefStore.refLocations.get(state.repo, state.rev, state.def) : null;
		state.authors = state.defObj ? DefStore.authors.get(state.repo, state.defObj.CommitID, state.def) : null;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.def !== prevState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantRefLocations(nextState.repo, nextState.rev, nextState.def));
		}

		if (prevState.defCommitID !== nextState.defCommitID && nextState.defCommitID) {
			if (ctx.features && ctx.features.Authors) {
				Dispatcher.Backends.dispatch(new DefActions.WantDefAuthors(nextState.repo, nextState.defCommitID, nextState.def));
			}
		}
	}

	render() {
		let def = this.state.defObj;
		let refLocs = this.state.refLocations;

		if (refLocs && refLocs.Error) {
			return (
				<Header
					title={`${httpStatusCode(refLocs.Error)}`}
					subtitle={`References are not available.`} />
			);
		}

		return (
			<div styleName="container">
				<div styleName="refs-page">
					<h1>{this.state.defObj && <Link to={urlToDef(this.state.defObj)}><code>{qualifiedNameAndType(this.state.defObj)}</code></Link>}</h1>
					{this.state.defObj && <p styleName="subheader">Defined in <Link to={urlToDef(this.state.defObj)}>{this.state.defObj.File}</Link></p>}
					<hr/>
					<div styleName="inner">
						<div styleName="def-info">
							<div styleName="main">
								{def && def.DocHTML && <p dangerouslySetInnerHTML={def && def.DocHTML}></p>}
								{def && !def.Error &&
									<div>
										<h2>Used in</h2>
										{!refLocs && <i>Loading...</i>}
										{refLocs && refLocs.filter((r) => r && r.Files).map((repoRef, i) => (
											<Link styleName="refs-link" key={i} to={urlToDefRefs(def, repoRef.Repo)}><span styleName="badge">{repoRef.Count}</span> {repoRef.Repo}</Link>
										))}
									</div>
								}
							</div>
							<div styleName="aside">
								{this.state.authors && <h2>Authors</h2>}
								{this.state.authors && <AuthorList authors={this.state.authors} />}
							</div>
						</div>
					</div>
				</div>
			</div>
		);
	}
}

export default CSSModules(DefInfo, styles);
