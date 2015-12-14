import React from "react";

import Component from "sourcegraph/Component";
import CodeFileRange from "../../../script/components/CodeFileRange"; // FIXME
import router from "../../../script/routing/router"; // FIXME

class TextSearchResults extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);

		if (props.resultData.Error) {
			state.error = props.resultData.Error;
			return;
		}

		state.results = props.resultData.Results;
		state.total = props.resultData.Total;
	}

	render() {
		if (this.state.error) {
			return <div className="alert alert-warning">There was an error returning your results: {this.state.error}</div>;
		}

		if (this.state.results.length === 0) {
			return <p className="summary">No text results found for "{this.state.query}"</p>;
		}
		let s = this.state.results.length === 1 ? "" : "s";
		let summary = `${this.state.total} text result${s} for "${this.state.query}"`;
		if (this.state.currentPage > 1) summary = `${summary} -- page ${this.state.currentPage}`;

		let currentFile, header;
		return (
			<div className="text-search-results">
				<p className="summary">{summary}</p>
				{this.state.results.map((result) => {
					if (currentFile !== result.File) {
						let fileURL = router.fileRangeURL(this.state.repo, this.state.rev, result.File, result.StartLine, this.state.results[this.state.results.length - 1].EndLine);
						header = <header><a href={fileURL}>{result.File}</a></header>;
					} else {
						header = null;
					}
					currentFile = result.File;

					return (
						<div className="text-search-result" key={`${result.File}-${result.StartLine}`}>
							{header}
							<CodeFileRange
								repo={this.state.repo}
								rev={this.state.rev}
								path={result.File}
								startLine={result.StartLine}
								endLine={result.EndLine}
								lines={result.Lines}
								showFileRangeLink={true} />
						</div>
					);
				})}
			</div>
		);
	}
}

TextSearchResults.propTypes = {
	repo: React.PropTypes.string,
	rev: React.PropTypes.string,
	query: React.PropTypes.string,
	resultData: React.PropTypes.object,
};

export default TextSearchResults;
