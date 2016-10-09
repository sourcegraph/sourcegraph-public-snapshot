// tslint:disable typedef ordered-imports
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
//
// The vscode editor JavaScript bundle can load either before or after
// the container element's ref callback has been called. We must
// handle both cases.
export class EditorComponent extends React.Component<Props, any> {

	private _container: HTMLElement;
	private _editor: Editor | null = null;

	componentWillMount(): void {
		require(["sourcegraph/editor/Editor"], ({Editor}) => {
			if (this._container) {
				this._editor = new Editor(this._container);
				if (this.props.editorRef) {
					this.props.editorRef(this._editor);
				}
			}
		});
	}

	componentWillUnmount(): void {
		if (this.props.editorRef) {
			this.props.editorRef(null);
		}
		if (this._editor) {
			this._editor.dispose();
		}
	}

	render(): JSX.Element {
		const otherProps = Object.assign({}, this.props);
		delete otherProps.editorRef;
		return <div ref={(e) => this.setContainer(e)} {...otherProps} />;
	}

	private setContainer(e: HTMLElement): void {
		this._container = e;

		if (e) {
			require(["sourcegraph/editor/Editor"], ({Editor}) => {
				if (!this._editor) {
					this._editor = new Editor(this._container);
					if (this.props.editorRef) {
						this.props.editorRef(this._editor);
					}
				}
			});
		} else if (this._editor) {
			this._editor.dispose();
			this._editor = null;
			if (this.props.editorRef) {
				this.props.editorRef(null);
			}
		}
	}
}
