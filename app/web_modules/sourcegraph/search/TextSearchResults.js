import React from "react";

import Blob from "sourcegraph/blob/Blob";
import Component from "sourcegraph/Component";
import {hotLinkAnyElement} from "sourcegraph/util/hotLink";
import * as router from "sourcegraph/util/router";

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
		if (this.state.currentPage > 1) summary = `Page ${this.state.currentPage} of ${summary}`;

		let currentFile;
		return (
			<div className="text-search-results theme-default">
				<p className="summary">{summary}</p>
				{this.state.results.map((result) => {
					const showHeader = currentFile !== result.File;
					currentFile = result.File;

					let url = router.tree(this.state.repo, this.state.rev, currentFile, result.StartLine, result.EndLine);

					return (
						<div className="text-search-result" key={`${result.File}-${result.StartLine}`}
							data-href={url} onClick={hotLinkAnyElement}>
							{showHeader ? <header><a href={url}>{result.File}</a></header> : null}
							<Blob
								repo={this.state.repo}
								rev={this.state.rev}
								path={currentFile}
								lineNumbers={true}
								startLine={result.StartLine}
								endLine={result.EndLine}
								contentsOffsetLine={result.StartLine}
								contents={result.Contents}
								highlightedDef={null}
								activeDef={null}
								annotations={queryHighlightAnnotations(this.state.query, result.Contents)} />
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

function queryHighlightAnnotations(query, match) {
	let indexes = [];
	let i = 0;
	while (i < match.length - query.length) {
		if (match.slice(i).startsWith(query)) {
			indexes.push(i);
			i += query.length;
		} else {
			i++;
		}
	}

	return indexes.map((j) => ({StartByte: j, EndByte: j + query.length, Class: "highlight-primary"}));
}
