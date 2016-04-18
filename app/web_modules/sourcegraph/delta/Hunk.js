import React from "react";

import {atob} from "abab";
import classNames from "classnames";
import BlobLine from "sourcegraph/blob/BlobLine";
import fileLines from "sourcegraph/util/fileLines";

class Hunk extends React.Component {
	render() {
		let hunk = this.props.hunk;

		let lines = fileLines(atob(hunk.Body));

		let origLine = hunk.OrigStartLine;
		let newLine = hunk.NewStartLine;

		let lineStartByte = 0;

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
						let thisLineStartByte = lineStartByte;
						lineStartByte += line.length + 1; // account for 1-char newline

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

						let anns = this.props.annotations ? this.props.annotations.filter((ann) => (
							ann.StartByte >= thisLineStartByte && ann.EndByte <= thisLineStartByte + line.length
						)) : null;

						return (
							<BlobLine
								key={i}
								className={classNames({
									"new-line": prefix === "+",
									"old-line": prefix === "-",
								})}
								oldLineNumber={prefix === "+" ? null : origLine}
								newLineNumber={prefix === "-" ? null : newLine}
								contents={line}
								startByte={thisLineStartByte}
								annotations={anns} />
						);
					})}
				</tbody>
			</table>
		);
	}
}
Hunk.propTypes = {
	hunk: React.PropTypes.object.isRequired,
	annotations: React.PropTypes.array,
};
export default Hunk;
