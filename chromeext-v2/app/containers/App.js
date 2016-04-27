import React, {Component, PropTypes} from "react";
import {formatPattern} from "react-router/lib/PatternUtils";

import {bindActionCreators} from "redux";
import {connect} from "react-redux";

import {qualifiedNameAndType} from "../components/Formatter";
import SearchMenu from "../components/SearchMenu";
import SearchInput from "../components/SearchInput";
import TextSearchResult from "../components/TextSearchResult";
import DefSearchResult from "../components/DefSearchResult";
import {keyFor} from "../reducers/helpers";
import * as Actions from "../actions";

import CSSModules from "react-css-modules";
import styles from "../components/App.css";

@connect(
	(state) => ({
		repo: state.repo,
		rev: state.rev,
		path: state.path,
		query: state.query,
		srclibDataVersion: state.srclibDataVersion,
		defs: state.defs,
		text: state.text,
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
		text: PropTypes.object.isRequired,
		actions: PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this.state = {
			selected: "def",
		}
	}

	handleSubmit = (query) => {
		this.props.actions.setQuery(query);
		this.props.actions.getDefs(this.props.repo, this.props.rev, this.props.path, query);
		this.props.actions.getText(this.props.repo, this.props.rev, this.props.path, query);
	};

	handleMenuSelect = (selected) => {
		this.setState({selected});
	}

	_srclibDataVersion() {
		return this.props.srclibDataVersion.content[keyFor(this.props.repo, this.props.rev, this.props.path)];
	}

	_defs() {
		const srclibDataVersion = this._srclibDataVersion();
		if (!srclibDataVersion || !srclibDataVersion.CommitID) return null;
		return this.props.defs.content[keyFor(this.props.repo, srclibDataVersion.CommitID, this.props.path, this.props.query)];
	}

	// TODO: share this code with main application.
	defPath(def) {
		return `${def.UnitType}/${def.Unit}/-/${def.Path}`;
	}

	defParams(def, rev) {
		rev = rev === null ? def.CommitID : rev;
		const revPart = rev ? `@${rev || def.CommitID}` : "";
		return {splat: [`${def.Repo}${revPart}`, this.defPath(def)]};
	}

	urlToDef(def, rev) {
		rev = rev === null ? def.CommitID : rev;
		if ((def.File === null || def.Kind === "package")) {
			// The def's File field refers to a directory (e.g., in the
			// case of a Go package). We can't show a dir in this view,
			// so just redirect to the dir listing.
			//
			// TODO(sqs): Improve handling of this case.
			// let file = def.File === "." ? "" : def.File;
			// return urlToTree(def.Repo, rev, file);
			console.log("TODO");
		}
		return formatPattern("*/-/def/*", this.defParams(def, rev));
	}

	_text() {
		const srclibDataVersion = this._srclibDataVersion();
		if (!srclibDataVersion || !srclibDataVersion.CommitID) return null;
		return this.props.text.content[keyFor(this.props.repo, srclibDataVersion.CommitID, this.props.path, this.props.query)];
	}

	render() {
		const srclibDataVersion = this._srclibDataVersion();
		const defs = this._defs();
		const text = this._text();
		return (
			<div styleName="app">
				<div className="column one-fourth">
					<SearchMenu onSelect={this.handleMenuSelect} selected={this.state.selected} />
				</div>
				<div className="column three-fourths">
					<div className="breadcrumb flex-table" styleName="input-box">
						<span styleName="input-addon">{`${this.props.repo.split('/')[2]} /`}</span>
						<SearchInput text={this.props.query} placeholder="Search for symbols..." onSubmit={this.handleSubmit} />
					</div>
					{this.state.selected === "def" && <div className="tree-finder clearfix" styleName="list">
						<table className="tree-browser css-truncate">
							<tbody className="tree-browser-result js-tree-browser-result">
							{defs && defs.Defs && defs.Defs.map((item, i) =>
								<DefSearchResult key={i} href={`https://sourcegraph.com/${this.urlToDef(item /*, rev */)}`} qualifiedNameAndType={qualifiedNameAndType(item)} />
							)}
							</tbody>
						</table>
						{!srclibDataVersion || !srclibDataVersion.CommitID &&
							<div styleName="list-item-empty">Fetching...</div>
						}
						{srclibDataVersion && srclibDataVersion.CommitID && (!defs || !defs.Defs || defs.Defs.length === 0) &&
							<div styleName="list-item-empty">No results.</div>
						}
					</div>}
					{this.state.selected === "text" && <div className="code-list" styleName="code-list">
						{text && text.SearchResults && text.SearchResults.map((item, i) =>
							<TextSearchResult key={i} query={this.props.query} match={item.Match} file={item.File} startLine={item.StartLine} endLine={item.EndLine} repo={this.props.repo}/>
						)}
						{srclibDataVersion && srclibDataVersion.CommitID && (!text || !text.SearchResults || text.SearchResults.length === 0) &&
							<div styleName="list-item-empty">No results.</div>
						}
					</div>
					}
				</div>
			</div>
		);
	}
}
