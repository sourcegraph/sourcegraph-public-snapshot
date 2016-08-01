import * as React from "react";
import {Link} from "react-router";
import CSSModules from "react-css-modules";
import Container from "sourcegraph/Container";
import DefStore from "sourcegraph/def/DefStore";
import s from "sourcegraph/def/styles/Def.css";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import {defPath} from "sourcegraph/def";
import * as DefActions from "sourcegraph/def/DefActions";
import Dispatcher from "sourcegraph/Dispatcher";
import RefLocationsList from "sourcegraph/def/RefLocationsList";
import AuthorList from "sourcegraph/def/AuthorList";
import {urlToDefInfo} from "sourcegraph/def/routes";

class DefPopup extends Container {
	static propTypes = {
		def: React.PropTypes.object.isRequired,
		rev: React.PropTypes.string,
		refLocations: React.PropTypes.object,
		path: React.PropTypes.string.isRequired,
		location: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		features: React.PropTypes.object.isRequired,
	};

	reconcileState(state, props) {
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

	render() {
		let def = this.props.def;
		let refLocs = this.props.refLocations;

		return (
			<div className={s.marginBox}>
				<header>
					<Link className={s.boxTitle} to={urlToDefInfo(this.state.defObj, this.state.rev)}><span styleName="def-title">{qualifiedNameAndType(def, {unqualifiedNameClass: s.defName})}</span></Link>
					<Link className={s.boxIcon} to={urlToDefInfo(this.state.defObj, this.state.rev)}>&raquo;</Link>
				</header>
				<Link to={urlToDefInfo(this.state.defObj, this.state.rev)}>
					<header className={s.sectionTitle}>Used in
						<span>
						{refLocs && refLocs.TotalRepos && ` ${refLocs.TotalRepos} repositor${refLocs.TotalRepos === 1 ? "y" : "ies"}`}
						{refLocs && !refLocs.TotalRepos && refLocs.RepoRefs && ` ${refLocs.RepoRefs.length}+ repositories`}
						</span>
					</header>
				</Link>

				{!refLocs && <span styleName="loading">Loading...</span>}
				{refLocs && (!refLocs.RepoRefs || refLocs.RepoRefs.length === 0) && <i>No usages found</i>}
				{<RefLocationsList def={def} refLocations={refLocs} showMax={3} repo={this.state.repo} rev={this.state.rev} path={this.state.path} location={this.state.location} />}

				{<header className={s.sectionTitle}>Authors</header>}
				{!this.state.authors && <span styleName="loading">Loading...</span>}
				{this.state.authors && !this.state.authors.Error && this.state.authors.DefAuthors.length && <AuthorList authors={this.state.authors.DefAuthors} />}
			</div>
		);
	}
}

export default CSSModules(DefPopup, s);
