import React, {Component, PropTypes} from "react";

import {bindActionCreators} from "redux";
import {connect} from "react-redux";

import {qualifiedNameAndType} from "../components/Formatter";
import SearchInput from "../components/SearchInput";
import DefSearchResult from "../components/DefSearchResult";
import {keyFor} from "../reducers/helpers";
import * as Actions from "../actions";

import _ from "lodash";

import CSSModules from "react-css-modules";
import styles from "../components/App.css";

import {default as checkErrorStatus} from "../actions/xhr"

@connect(
	(state) => ({
		repo: state.repo,
		rev: state.rev,
		path: state.path,
		query: state.query,
		srclibDataVersion: state.srclibDataVersion,
		defs: state.defs,
	}),
	(dispatch) => ({
		actions: bindActionCreators(Actions, dispatch)
	})
)
@CSSModules(styles)
export default class App extends Component {
	static propTypes = {
		repo: PropTypes.string.isRequired,
		rev: PropTypes.string.isRequired,
		path: PropTypes.string,
		query: PropTypes.string.isRequired,
		srclibDataVersion: PropTypes.object.isRequired,
		defs: PropTypes.object.isRequired,
		actions: PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this.handleSubmit = _.debounce(this.handleSubmit, 50);
		this.handleSubmit = this.handleSubmit.bind(this);
	}

	handleSubmit = (query) => {
		this.props.actions.setQuery(query);
		this.props.actions.getDefs(this.props.repo, this.props.rev, this.props.path, query);
	};

	_srclibDataVersion() {
		return this.props.srclibDataVersion.content[keyFor(this.props.repo, this.props.rev, this.props.path)];
	}

	_srclibDataVersionFetch() {
		return this.props.srclibDataVersion.fetches[keyFor(this.props.repo, this.props.rev, this.props.path)];
	}

	_defs() {
		const srclibDataVersion = this._srclibDataVersion();
		if (!srclibDataVersion || !srclibDataVersion.CommitID) return null;
		return this.props.defs.content[keyFor(this.props.repo, srclibDataVersion.CommitID, this.props.path, this.props.query)];
	}

	_defsFetch() {
		const srclibDataVersion = this._srclibDataVersion();
		if (!srclibDataVersion || !srclibDataVersion.CommitID) return null;
		return this.props.defs.fetches[keyFor(this.props.repo, srclibDataVersion.CommitID, this.props.path, this.props.query)];
	}

	// TODO: Remove this, re-use this instead:
	//
	//       	import {urlToDefInfo} from "sourcegraph/def/routes";
	//
	//       Need to verify it doesn't break the DefSearchResult component below that uses urlToDef.
	//       Who's familiar with it?
	urlToDef(def, rev) {
		rev = rev ? rev : (def.CommitID || "");
		return `${def.Repo}${rev ? `@${rev}` : ""}/-/info/${def.UnitType}/${def.Unit}/-/${def.Path}`;
	}

	render() {
		const srclibDataVersion = this._srclibDataVersion();
		const srclibDataVersionFetch = this._srclibDataVersionFetch();
		const defs = this._defs();
		const defsFetch = this._defsFetch();
		return (
			<div styleName="app">
				<div styleName="full-column">
					<div className="breadcrumb flex-table" styleName="input-box">
						<span styleName="input-addon">{`${this.props.repo.split('/')[2]} /`}</span>
						<SearchInput placeholder="Search for symbols..." onSubmit={this.handleSubmit} onChange={this.handleSubmit} />
					</div>
					<div className="tree-finder clearfix" styleName="list">
						<table className="tree-browser css-truncate">
							<tbody className="tree-browser-result js-tree-browser-result">
							{defs && defs.Defs && defs.Defs.map((item, i) =>
								<DefSearchResult key={i} href={`https://sourcegraph.com/${this.urlToDef(item, this.props.rev)}`} qualifiedNameAndType={qualifiedNameAndType(item)} />
							)}
							</tbody>
						</table>
						{!srclibDataVersion || !srclibDataVersion.CommitID &&
							<h3 styleName="list-item-empty">Fetching...</h3>
						}
						{srclibDataVersionFetch === true || defsFetch === true &&
							<h3 styleName="list-item-empty">Searching...</h3>
						}
						{srclibDataVersionFetch && srclibDataVersionFetch.response && srclibDataVersionFetch.response.status === 404 &&
							<h3 styleName="list-item-empty">404 Not Found: This repository (or revision) has not been indexed by Sourcegraph.<br/><a href={`https://sourcegraph.com/${this.props.repo}@${this.props.rev}`}>Let Sourcegraph index this repository.</a></h3>
						}
						{srclibDataVersionFetch && srclibDataVersionFetch.response && srclibDataVersionFetch.response.status === 401 &&
							<h3 styleName="list-item-empty">401 Unauthorized: Log in to Sourcegraph to search private code</h3>
						}
						{srclibDataVersionFetch === false && defs && !defs.Defs &&
							<h3 styleName="list-item-empty">No matches found.</h3>
						}
					</div>
				</div>
			</div>
		);
	}
}
