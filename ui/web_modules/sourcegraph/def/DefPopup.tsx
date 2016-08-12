// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";
import {Container} from "sourcegraph/Container";
import {DefStore} from "sourcegraph/def/DefStore";
import * as styles from "sourcegraph/def/styles/Def.css";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import {defPath} from "sourcegraph/def/index";
import * as DefActions from "sourcegraph/def/DefActions";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {RefLocationsList} from "sourcegraph/def/RefLocationsList";
import {LocalRefLocationsList} from "sourcegraph/def/LocalRefLocationsList";
import {AuthorList} from "sourcegraph/def/AuthorList";
import {urlToDefInfo} from "sourcegraph/def/routes";

interface Props {
	def: any;
	rev?: string;
	refLocations?: any;
	path?: string;
	location?: any;
}

export class DefPopup extends Container<Props, any> {
	static contextTypes = {
		features: React.PropTypes.object.isRequired,
	};

	reconcileState(state, props: Props) {
		Object.assign(state, props);
		state.defObj = props.def;
		state.repo = props.def ? props.def.Repo : null;
		state.rev = props.rev || null;
		state.commitID = props.def ? props.def.CommitID : null;
		state.def = props.def ? defPath(props.def) : null;

		state.authors = DefStore.authors.get(state.repo, state.commitID, state.def);
	}

	onStateTransition(prevState, nextState) {
		if (prevState.repo !== nextState.repo || prevState.commitID !== nextState.commitID || prevState.def !== nextState.def) {
			Dispatcher.Backends.dispatch(new DefActions.WantDefAuthors(nextState.repo, nextState.commitID, nextState.def));
		}
	}

	stores() { return [DefStore]; }

	render(): JSX.Element | null {
		let def = this.props.def;
		let refLocs = this.props.refLocations;

		return (
			<div className={styles.marginBox}>
				<header>
					<Link className={styles.boxTitle} to={urlToDefInfo(this.state.defObj, this.state.rev)}><span className={styles.def_title}>{qualifiedNameAndType(def, {unqualifiedNameClass: styles.defName})}</span></Link>
					<Link className={styles.boxIcon} to={urlToDefInfo(this.state.defObj, this.state.rev)}>&raquo;</Link>
				</header>
				<Link to={urlToDefInfo(this.state.defObj, this.state.rev)}>
					<header className={styles.sectionTitle}>Used in
						<span>
						{refLocs && refLocs.TotalRepos && ` ${refLocs.TotalRepos} repositor${refLocs.TotalRepos === 1 ? "y" : "ies"}`}
						{refLocs && !refLocs.TotalRepos && refLocs.RepoRefs && ` ${refLocs.RepoRefs.length}+ repositories`}
						{refLocs && refLocs.TotalFiles && ` ${refLocs.TotalFiles} file${refLocs.TotalFiles === 1 ? "" : "s"}`}
						</span>
					</header>
				</Link>

				{!refLocs && <span className={styles.loading}>Loading...</span>}
				{refLocs && (!refLocs.RepoRefs || refLocs.RepoRefs.length === 0) && (!refLocs.Files || refLocs.Files.length === 0) && <i>No usages found</i>}
				{<RefLocationsList def={def} refLocations={refLocs} showMax={3} repo={this.state.repo} rev={this.state.rev} path={this.state.path} location={this.state.location} />}
				{refLocs && refLocs.RepoRefs &&
					<RefLocationsList def={def} refLocations={refLocs} showMax={3} repo={this.state.repo} rev={this.state.rev} path={this.state.path} location={this.state.location} />}
				{refLocs && refLocs.Files &&
					<LocalRefLocationsList refLocations={refLocs} showMax={5} repo={this.state.repo} rev={this.state.rev} path={this.state.path} />}

				{<header className={styles.sectionTitle}>Authors</header>}
				{!this.state.authors && <span className={styles.loading}>Loading...</span>}
				{this.state.authors && !this.state.authors.Error && this.state.authors.DefAuthors.length && <AuthorList authors={this.state.authors.DefAuthors} />}
			</div>
		);
	}
}
