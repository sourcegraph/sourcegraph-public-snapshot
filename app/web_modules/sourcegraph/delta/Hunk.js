import React from "react";

import {atob} from "abab";
import classNames from "classnames";

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
							<tr key={i} className={classNames({
								"line": true,
								"new-line": prefix === "+",
								"old-line": prefix === "-",
							})}>
								<td className="line-number" data-line={prefix === "+" ? "" : origLine}></td>
								<td className="line-number" data-line={prefix === "-" ? "" : newLine}></td>
								<td className="line-content">
									<span className="prefix">{prefix}</span>
									{line.slice(1)}
								</td>
							</tr>
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
