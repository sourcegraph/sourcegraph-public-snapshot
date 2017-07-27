import { fetchBlobContent } from "app/backend";
import { getPathExtension, getModeFromExtension } from "app/util";
import * as colors from "app/util/colors";
import { highlightNode } from "app/util/dom";
import { BlobPosition } from "app/util/types";
import { highlightBlock } from "highlight.js";
import * as React from "react";
import { classes, style } from "typestyle";

interface Props extends BlobPosition {
	// How many extra lines to show in the excerpt before/after the ref.
	previewWindowExtraLines?: number;
	highlightLength: number;
}

interface State {
	blobLines?: string[];
}

namespace Styles {
	export const lineNum = style({ color: colors.baseColor, paddingRight: "15px" });
	export const codeLine = style({ whiteSpace: "pre" });
}

export class CodeExcerpt extends React.Component<Props, State> {
	constructor(props: Props) {
		super(props);
		this.state = {};
	}

	componentDidMount(): void {
		fetchBlobContent(this.props.uri, this.props.rev, this.props.path).then(content => {
			const blobLines = content.split("\n");
			this.setState({ blobLines });
		});
	}

	getPreviewWindowLines(): number[] {
		const targetLine = this.props.line;
		let res = [targetLine];
		for (let i = targetLine - this.props.previewWindowExtraLines!; i < targetLine + this.props.previewWindowExtraLines! + 1; ++i) {
			if (i > 0 && i < targetLine) {
				res = [i].concat(res);
			}
			if (i < this.state.blobLines!.length && i > targetLine) {
				res = res.concat([i]);
			}
		}
		return res;
	}

	render(): JSX.Element | null {
		if (!this.state.blobLines) {
			return null;
		}
		return <table>
			<tbody>
				{this.getPreviewWindowLines().map(i => {
					return <tr key={i}>
						<td className={Styles.lineNum}>{i}</td>
						<td className={classes(getModeFromExtension(getPathExtension(this.props.path)), Styles.codeLine)} ref={(el) => {
							if (el) {
								highlightBlock(el);
								if (i === this.props.line) {
									highlightNode(el, this.props.char!, this.props.highlightLength);
								}
							}
						}}>{this.state.blobLines![i]}</td>
					</tr>;
				})}
			</tbody>
		</table>;
	}
}
