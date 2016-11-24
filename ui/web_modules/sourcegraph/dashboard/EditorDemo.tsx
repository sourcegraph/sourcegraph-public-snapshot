import * as React from "react";
import { URIUtils } from "sourcegraph/core/uri";
import { Editor } from "sourcegraph/editor/Editor";
import { EditorComponent } from "sourcegraph/editor/EditorComponent";
import { Range } from "vs/editor/common/core/range";

type Props = {
	repo: string;
	rev: string;
	path: string;
	startLine: number;
};
type State = Props & {
	contents: string | null;
};

// The purpose of this file is to easily render one example during the
// onboarding flow that will allow the user to hover over examples in
// a sandboxed mode. This is similar to the regular RefsContainer
// except that css styles and eventlisteners / additional fetches for
// information are removed.
export class EditorDemo extends React.Component<Props, State> {
	constructor(props: Props) {
		super(props);
		this._setEditor = this._setEditor.bind(this);
	}

	private _setEditor(editor: Editor): void {
		if (editor) {
			const range = new Range(this.props.startLine, 1, this.props.startLine, 1);
			const uri = URIUtils.pathInRepo(this.props.repo, this.props.rev, this.props.path);
			editor.setInput(uri, range);
		}
	}

	render(): JSX.Element | null {
		return (
			<EditorComponent editorRef={this._setEditor} style={{ display: "flex", flexDirection: "column", height: "225px", border: "solid 1px #efefef" }} />
		);
	}
}
