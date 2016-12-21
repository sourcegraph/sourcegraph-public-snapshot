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

	// _prevContainer saves the previous HTMLElement so we can reuse
	// the editor instance if the HTML element is reused. This
	// improves perf (esp. of jumping between files).
	private _prevContainer: HTMLElement | undefined;

	componentWillUnmount(): void {
		if (this._editor) {
			this._editor.dispose();
		}
	}

	render(): JSX.Element {
		return <div ref={(e) => this.setContainer(e)} style={this.props.style} />;
	}

	private setContainer(container: HTMLElement): void {
		if (container) {
			this._isMounted = true;

			// Reuse the previous editor instance if we're reusing the
			// same underlying element.
			if (this._prevContainer === container) {
				if (this.props.editorRef) {
					this.props.editorRef(this._editor);
				}
				return;
			}

			// Otherwise dispose the editor now (we did not do that
			// when unmounted, to preserve the opportunity to reuse
			// it) and zero out our reference to the container (to
			// avoid a memory leak).
			if (this._prevContainer) {
				this._prevContainer = undefined;
			}
			if (this._editor) {
				this._editor.dispose();
			}

			this._prevContainer = container;
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
			if (this.props.editorRef) {
				this.props.editorRef(null);
			}
		}
	}
}
