// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "sourcegraph/blob/styles/Blob.css";
import * as classNames from "classnames";

const dummyLineLengths = [
	15, 0, 8, 10, 0, 28, 34, 0, 41, 0, 30,
	66, 72, 76, 65, 66, 58, 51, 46, 55, 2,
	0, 62, 0, 25, 0, 90, 39, 25, 30, 5, 0,
	67, 23, 5, 0, 72, 25, 35, 39, 62, 5,
	25, 27, 32, 24, 29, 31, 37, 51, 35, 42,
];

interface Props {
	styles?: any;
	numLines?: number;
}

// BlobContentPlaceholder implements the "content placeholder" effect
// seen in the loading screens of mobile apps and the Facebook news
// feed, where the initial state mimics the layout of what will be
// eventually displayed.
//
// It makes the app feel a lot less jittery.
export function BlobContentPlaceholder(props: Props) {
	const numLines = props.numLines || 60;
	const lines: any[] = [];
	for (let i = 0; i < numLines; i++) {
		const line = i + 1;
		lines.push(
			<tr className={styles.line} data-line={line} key={i}>
				<td className={styles.lineNumberCell}></td>
				<td className={classNames("code", styles.lineContentPlaceholder)} data-line={line}>
					<div className={styles.placeholderWhitespace} style={{width: `${100 - dummyLineLengths[i % dummyLineLengths.length]}%`}}>&nbsp;</div>
				</td>
			</tr>
		);
	}

	return (
		<div className={styles.scroller}>
			<table className={styles.linesContentPlaceholder}>
				<tbody>{lines}</tbody>
			</table>
		</div>
	);
}
