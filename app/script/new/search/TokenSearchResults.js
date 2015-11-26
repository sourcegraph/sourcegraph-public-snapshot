import React from "react";

import Component from "../Component";
import router from "../../routing/router"; // FIXME

const helpDocURL = "https://src.sourcegraph.com/sourcegraph/.docs/troubleshooting/builds/";

class TokenSearchResultsView extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);

		state.results = props.resultData.Results;
		state.total = props.resultData.Total;
		state.buildInfo = props.resultData.BuildInfo;
	}

	render() {
		let summary;
		if (this.state.results.length === 0) {
			summary = `No definition results found for "${this.state.query}"`;
		} else {
			let s = this.state.results.length === 1 ? "" : "s";
			summary = `${this.state.total} definition result${s} for "${this.state.query}"`;
			if (this.state.currentPage > 1) summary = `Page ${this.state.currentPage} of ${summary}`;
		}

		return (
			<div className="token-search-results">
				{!this.state.buildInfo &&
					<div className="alert alert-info">
						<i className="fa fa-warning"></i>No Code Intelligence data for {this.state.repo}. <a href={helpDocURL}>See troubleshooting guide</a>.
					</div>
				}
				{(this.state.buildInfo && !this.state.buildInfo.Exact) &&
					<div className="alert alert-info">
						<i className="fa fa-warning"></i>Showing definition results from {this.state.buildInfo.CommitsBehind} commit{this.state.buildInfo.CommitsBehind === 1 ? "" : "s"} behind latest. Newer results are shown when available.
					</div>
				}
				<p className="summary">{summary}</p>
				{this.state.results.map((result) => {
					let doc;
					if (result.Def.DocHTML) {
						// This HTML should be sanitized in ui/search.go
						doc = <p dangerouslySetInnerHTML={result.Def.DocHTML}></p>;
					}
					let def = result.Def;
					let href = router.defURL(def.Repo, def.CommitID, def.UnitType, def.Unit, def.Path);
					return (
						<div key={result.URL} className="token-search-result">
							<hr/>
							<a href={href}>
								<code>{result.Def.Kind} </code>
								{/* This HTML should be sanitized in ui/search.go */}
								<code dangerouslySetInnerHTML={result.QualifiedName}></code>
							</a>
							{doc}
						</div>
					);
				})}
				<div className="search-pagination">
				</div>
			</div>
		);
	}
}

TokenSearchResultsView.propTypes = {
	repo: React.PropTypes.string,
	resultData: React.PropTypes.object,
};

module.exports = TokenSearchResultsView;
