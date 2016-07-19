import React from "react";
import Component from "sourcegraph/Component";
import styles from "sourcegraph/delta/styles/Hunk.css";
import CSSModules from "react-css-modules";
import {atob} from "abab";
import BlobLine from "sourcegraph/blob/BlobLine";
import fileLines from "sourcegraph/util/fileLines";

class Hunk extends Component {
	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		let hunk = this.props.hunk;

		let lines = fileLines(atob(hunk.Body));

		let origLine = hunk.OrigStartLine || 0;
		let newLine = hunk.NewStartLine || 0;

		let lineStartByte = 0;

		return (
			<table styleName="lines">
				<tbody>
					<tr styleName="line">
						<td styleName="line-number-cell"><span styleName="line-number">...</span></td>
						<td styleName="line-number-cell"><span styleName="line-number">...</span></td>
						<td styleName="line-content">
							@@ -{hunk.OrigStartLine},{hunk.OrigLines} +{hunk.NewStartLine},{hunk.NewLines} @@ {hunk.Section}
						</td>
					</tr>

					{lines.map((line, i) => {
						let thisLineStartByte = lineStartByte;
						lineStartByte += line.length + 1; // account for 1-char newline

						const prefix = line[0];
						let styleName;
						if (i > 0) {
							switch (prefix) {
							case "+":
								newLine++;
								styleName = styles["new-line"];
								break;
							case "-":
								origLine++;
								styleName = styles["old-line"];
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
								className={styleName || null}
								oldLineNumber={prefix === "+" ? null : origLine}
								newLineNumber={prefix === "-" ? null : newLine}
								contents={line}
								startByte={thisLineStartByte}
								highlightedDef={this.props.highlightedDef}
								highlightedDefObj={this.props.highlightedDefObj}
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

	highlightedDef: React.PropTypes.string,
	highlightedDefObj: React.PropTypes.object,
};
export default CSSModules(Hunk, styles);
