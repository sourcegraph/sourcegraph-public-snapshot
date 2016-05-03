// @flow weak

import React from "react";
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
import {urlToDef} from "sourcegraph/def/routes";

class DefPopup extends Container {
	static propTypes = {
		def: React.PropTypes.object.isRequired,
		refLocations: React.PropTypes.array,
		path: React.PropTypes.string.isRequired,
	};

	static contextTypes = {
		features: React.PropTypes.object.isRequired,
	};

	reconcileState(state, props) {
		Object.assign(state, props);
		state.defObj = props.def;
		state.repo = props.def ? props.def.Repo : null;
		state.rev = props.def ? props.def.CommitID : null;
		state.def = props.def ? defPath(props.def) : null;

		state.authors = DefStore.authors.get(state.repo, state.rev, state.def);
	}

	onStateTransition(prevState, nextState) {
		if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.def !== nextState.def) {
			if (this.context.features.Authors) {
				Dispatcher.Backends.dispatch(new DefActions.WantDefAuthors(nextState.repo, nextState.rev, nextState.def));
			}
		}
	}

	stores() { return [DefStore]; }

	render() {
		let def = this.props.def;
		let refLocs = this.props.refLocations;

		return (
			<div className={s.marginBox}>
				<header className={s.boxTitle}>
					<Link to={`${urlToDef(this.state.defObj, this.state.rev)}/-/info`}><span styleName="def-title">{qualifiedNameAndType(def, {unqualifiedNameClass: s.defName})}</span></Link>
				</header>
				<header className={s.sectionTitle}>Used in</header>

				{!refLocs && <span styleName="loading">Loading...</span>}
				{refLocs && refLocs.length === 0 && <i>No usages found</i>}
				{<RefLocationsList def={def} refLocations={refLocs} repo={this.state.repo} rev={this.state.rev} path={this.state.path} />}

				{this.state.authors && <header className={s.sectionTitle}>Authors</header>}
				{!this.state.authors && <span styleName="loading">Loading...</span>}
				{this.state.authors && <AuthorList authors={this.state.authors} />}
			</div>
		);
	}
}

export default CSSModules(DefPopup, s);
