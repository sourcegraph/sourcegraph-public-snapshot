import React from "react";

import CodeLineView from "./CodeLineView";

class CodeListing extends React.Component {
	render() {
		// TODO tiled rendering for better performance on huge files
		return (
			<div className="code-view-react">
				<table className="line-numbered-code">
					<tbody>
						{this.props.lines.map((lineData, i) => <CodeLineView lineNumber={i + 1} tokens={lineData.Tokens || []} key={i} />)}
					</tbody>
				</table>
			</div>
		);
	}
}

CodeListing.propTypes = {
	lines: React.PropTypes.array,
};

export default CodeListing;
