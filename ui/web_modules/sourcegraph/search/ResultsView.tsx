import * as React from "react";
import { Link } from "react-router";

import URI from "vs/base/common/uri";

import { urlToBlob } from "sourcegraph/blob/routes";
import { getURIContext } from "sourcegraph/workbench/utils";

interface Props {
	results?: GQL.ISearchResults;
	loading: boolean;
}
const resultsSx = {
	margin: "0 auto",
	maxWidth: 800,
};

export function ResultsView(props: Props): JSX.Element {
	if (props.loading) {
		return <div style={resultsSx}>
			Loading
		</div>;
	}

	if (!props.results) {
		return <div style={resultsSx}>
			Enter a query.
		</div>;
	}
	if (props.results.results.length === 0) {
		return <div style={resultsSx}>
			No results found.
		</div>;
	}
	return <div style={resultsSx}>
		Here are some results:
		{props.results.results.map(FileResult)}
	</div>;
}

function FileResult(fileMatch: GQL.IFileMatch, key: number): JSX.Element {
	return <div key={key}>
		<File resource={fileMatch.resource} />
		{fileMatch.lineMatches.map(LineMatch)}
	</div>;
}

function File(props: { resource: string }): JSX.Element {
	const { repo, rev, path } = getURIContext(URI.parse(props.resource));

	return <Link to={urlToBlob(repo, rev, path)}>
		{repo} -- {path}
	</Link>;
}

function LineMatch(match: GQL.ILineMatch, key: number): JSX.Element {
	return <div key={key}>
		{match.lineNumber}: {match.preview}
	</div>;
}
