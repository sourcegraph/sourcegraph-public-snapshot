import React from "react";

import Component from "sourcegraph/Component";
import router from "../../../script/routing/router"; // FIXME

const helpDocURL = "https://src.sourcegraph.com/sourcegraph/.docs/troubleshooting/builds/";

class TokenSearchResultsView extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);

		if (props.resultData.Error) {
			state.error = props.resultData.Error;
			return;
		}

		state.results = props.resultData.Results;
		state.total = props.resultData.Total;
		state.srclibDataVersion = props.resultData.SrclibDataVersion;
	}

	render() {
		if (this.state.error) {
			return <div className="alert alert-warning">There was an error returning your results: {this.state.error}</div>;
		}

		let summary;
		if (this.state.results.length === 0) {
			summary = `No definition results found for "${this.state.query}"`;
		} else {
			let s = this.state.results.length === 1 ? "" : "s";
			summary = `${this.state.total} definition result${s} for "${this.state.query}"`;
			if (this.state.currentPage > 1) summary = `${summary} -- page ${this.state.currentPage}`;
		}

		return (
			<div className="token-search-results">
				{!this.state.srclibDataVersion &&
					<div className="alert alert-info">
						<i className="fa fa-warning"></i>No Code Intelligence data for {this.state.repo}. <a href={helpDocURL}>See troubleshooting guide</a>.
					</div>
				}
				{(this.state.srclibDataVersion && this.state.srclibDataVersion.CommitsBehind) &&
					<div className="alert alert-info">
						<i className="fa fa-warning"></i>Showing definition results from {this.state.srclibDataVersion.CommitsBehind} commit{this.state.srclibDataVersion.CommitsBehind === 1 ? "" : "s"} behind latest. Newer results are shown when available.
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
			</div>
		);
	}
}

TokenSearchResultsView.propTypes = {
	repo: React.PropTypes.string,
	query: React.PropTypes.string,
	resultData: React.PropTypes.object,
};

export default TokenSearchResultsView;
