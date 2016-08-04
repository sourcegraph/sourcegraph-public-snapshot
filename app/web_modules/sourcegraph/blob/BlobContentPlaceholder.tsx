// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import styles from "sourcegraph/blob/styles/Blob.css";

const dummyLineLengths = [
	15, 0, 8, 10, 0, 28, 34, 0, 41, 0, 30,
	66, 72, 76, 65, 66, 58, 51, 46, 55, 2,
	0, 62, 0, 25, 0, 90, 39, 25, 30, 5, 0,
	67, 23, 5, 0, 72, 25, 35, 39, 62, 5,
	25, 27, 32, 24, 29, 31, 37, 51, 35, 42,
];

// BlobContentPlaceholder implements the "content placeholder" effect
// seen in the loading screens of mobile apps and the Facebook news
// feed, where the initial state mimics the layout of what will be
// eventually displayed.
//
// It makes the app feel a lot less jittery.
function BlobContentPlaceholder(props) {
	let s = props.styles;

	const numLines = props.numLines || 60;
	const lines: any[] = [];
	for (let i = 0; i < numLines; i++) {
		const line = i + 1;
		lines.push(
			<tr className={s.line} data-line={line} key={i}>
				<td className={s.lineNumberCell}></td>
				<td className={`code ${s.lineContentPlaceholder}`} data-line={line}>
					<div className={s.placeholderWhitespace} style={{width: `${100 - dummyLineLengths[i % dummyLineLengths.length]}%`}}>&nbsp;</div>
				</td>
			</tr>
		);
	}

	return (
		<div className={s.scroller}>
			<table className={s.linesContentPlaceholder}>
				<tbody>{lines}</tbody>
			</table>
		</div>
	);
}

(BlobContentPlaceholder as any).propTypes = {
	styles: React.PropTypes.object,
	numLines: React.PropTypes.number,
};

export default CSSModules(BlobContentPlaceholder, styles);
