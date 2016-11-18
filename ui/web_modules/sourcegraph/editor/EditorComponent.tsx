import * as React from "react";
import "sourcegraph/blob/styles/Monaco.css";

type Props = {
	editorRef?: (editor: Editor | null) => void;
	style?: Object;
};

interface Editor {
	dispose(): void;
} // avoid importing it so we can use code splitting

// EditorComponent wraps the Monaco loader and code editor. Embedders
// must use the editorRef property to get the Monaco editor object and
// manipulate it using the Monaco API.
export class EditorComponent extends React.Component<Props, any> {
	private _isMounted: boolean = false;
	private _editor: Editor | null = null;

	render(): JSX.Element {
		return <div ref={(e) => this.setContainer(e)} style={this.props.style} />;
	}

	private setContainer(container: HTMLElement): void {
		if (container) {
			this._isMounted = true;
			require(["sourcegraph/editor/Editor"], ({Editor}) => {
				if (!this._isMounted) { // component got unmounted before "require" finished
					return;
				}
				this._editor = new Editor(container);
				if (this.props.editorRef) {
					this.props.editorRef(this._editor);
				}
			});
		} else {
			this._isMounted = false;
			if (!this._editor) { // component got unmounted before "require" finished
				return;
			}
			this._editor.dispose();
			if (this.props.editorRef) {
				this.props.editorRef(null);
			}
		}
	}
}
