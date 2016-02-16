import React from "react";

import {atob} from "abab";
import classNames from "classnames";
import CodeLineView from "sourcegraph/code/CodeLineView";

class Hunk extends React.Component {
	render() {
		let hunk = this.props.hunk;

		let lines = atob(hunk.Body).split("\n");
		if (lines.length > 0 && lines[lines.length - 1] === "") {
			lines.pop();
		}

		let origLine = hunk.OrigStartLine;
		let newLine = hunk.NewStartLine;

		return (
			<table className="line-numbered-code theme-default file-diff-hunk">
				<tbody>
					<tr className="line hunk-header">
						<td className="line-number">...</td>
						<td className="line-number">...</td>
						<td className="line-content">
							@@ -{hunk.OrigStartLine},{hunk.OrigLines} +{hunk.NewStartLine},{hunk.NewLines} @@ {hunk.Section}
						</td>
					</tr>

					{lines.map((line, i) => {
						const prefix = line[0];
						if (i > 0) {
							switch (prefix) {
							case "+":
								newLine++;
								break;
							case "-":
								origLine++;
								break;
							case " ":
								origLine++;
								newLine++;
								break;
							}
						}
						return (
							<CodeLineView
								key={i}
								className={classNames({
									"new-line": prefix === "+",
									"old-line": prefix === "-",
								})}
								oldLineNumber={prefix === "+" ? null : origLine}
								newLineNumber={prefix === "-" ? null : newLine}
								contents={line} />
						);
					})}
				</tbody>
			</table>
		);
	}
}
Hunk.propTypes = {
	hunk: React.PropTypes.object.isRequired,
};
export default Hunk;
