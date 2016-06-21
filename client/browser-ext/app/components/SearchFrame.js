import React from "react";

import {bindActionCreators} from "redux";
import {connect} from "react-redux";

import {qualifiedNameAndType} from "./Formatter";
import SearchInput from "./SearchInput";
import DefSearchResult from "./DefSearchResult";
import {keyFor} from "../reducers/helpers";
import * as Actions from "../actions";
import * as utils from "../utils";

import _ from "lodash";

import CSSModules from "react-css-modules";
import styles from "./App.css";

@connect(
	(state) => ({
		resolvedRev: state.resolvedRev,
		srclibDataVersion: state.srclibDataVersion,
		defs: state.defs,
	}),
	(dispatch) => ({
		actions: bindActionCreators(Actions, dispatch)
	})
)
@CSSModules(styles)
export default class SearchFrame extends React.Component {
	static propTypes = {
		resolvedRev: React.PropTypes.object.isRequired,
		srclibDataVersion: React.PropTypes.object.isRequired,
		defs: React.PropTypes.object.isRequired,
		actions: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this._handleSubmit = _.debounce(this._handleSubmit, 50);
		this._handleSubmit = this._handleSubmit.bind(this);
		this._refresh = this._refresh.bind(this);

		this.state = utils.parseURL();
		this.state.query = "";
	}

	componentDidMount() {
		document.addEventListener("pjax:success", this._refresh);
	}

	componentWillUnmount() {
		document.removeEventListener("pjax:success", this._refresh);
	}

	_refresh() {
		const newState = utils.parseURL();
		if (newState.repoURI !== this.state.repoURI) {
			newState.query = "";
		} else {
			newState.query = this.state.query;
		}
		this.setState(newState);
	}

	_handleSubmit(query) {
		this.props.actions.getDefs(this.state.repoURI, this.state.rev, this.state.path, query);
		this.setState({query});
	};

	_srclibDataVersion() {
		const resolvedRev = this.props.resolvedRev.content[keyFor(this.state.repoURI, this.state.rev)];
		if (!resolvedRev) return null;

		return this.props.srclibDataVersion.content[keyFor(this.state.repoURI, resolvedRev.CommitID, this.state.path)];
	}

	_srclibDataVersionFetch() {
		const resolvedRev = this.props.resolvedRev.content[keyFor(this.state.repoURI, this.state.rev)];
		if (!resolvedRev) return null;

		return this.props.srclibDataVersion.fetches[keyFor(this.state.repoURI, resolvedRev.CommitID, this.state.path)];
	}

	_defs() {
		const srclibDataVersion = this._srclibDataVersion();
		if (!srclibDataVersion || !srclibDataVersion.CommitID) return null;
		return this.props.defs.content[keyFor(this.state.repoURI, srclibDataVersion.CommitID, this.state.path, this.state.query)];
	}

	_defsFetch() {
		const srclibDataVersion = this._srclibDataVersion();
		if (!srclibDataVersion || !srclibDataVersion.CommitID) return null;
		return this.props.defs.fetches[keyFor(this.state.repoURI, srclibDataVersion.CommitID, this.state.path, this.state.query)];
	}

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
						<span styleName="input-addon">{`${this.state.repoURI.split('/')[2]} /`}</span>
						<SearchInput placeholder="Search for symbols..." onSubmit={this._handleSubmit} onChange={this._handleSubmit} />
					</div>
					<div className="tree-finder clearfix" styleName="list">
						<table className="tree-browser css-truncate">
							<tbody className="tree-browser-result js-tree-browser-result">
							{defs && defs.Defs && defs.Defs.map((item, i) =>
								<DefSearchResult key={i} href={`https://sourcegraph.com/${this.urlToDef(item, this.state.rev)}`} query={this.state.query} qualifiedNameAndType={qualifiedNameAndType(item)} />
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
							<h3 styleName="list-item-empty">404 Not Found: This repository (or revision) has not been indexed by Sourcegraph.<br/><a href={`https://sourcegraph.com/${this.state.repoURI}@${this.state.rev}`}>Let Sourcegraph index this repository.</a></h3>
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
