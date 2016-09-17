// tslint:disable typedef ordered-imports
import * as React from "react";
import * as invariant from "invariant";
import {loadMonaco} from "sourcegraph/editor/loader";
import {Editor} from "sourcegraph/editor/Editor";
import "sourcegraph/blob/styles/Monaco.css";

type Props = {
	editorRef?: (editor: Editor | null) => void;
	style?: Object;
};

interface State {
	monacoLoaded?: boolean;
};

// EditorComponent wraps the Monaco loader and code editor. Embedders
// must use the editorRef property to get the Monaco editor object and
// manipulate it using the Monaco API.
export class EditorComponent extends React.Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
	};

	context: {
		siteConfig: { assetsRoot: string };
	};

	private _container: HTMLElement;
	private _editor?: Editor;

	componentWillMount(): void {
		loadMonaco(this.context.siteConfig.assetsRoot).then(() => {
			invariant(!this._editor, "editor is already initialized");
			invariant(this._container, "container element ref is not available");

			this._editor = new Editor(this._container);

			if (this.props.editorRef) {
				this.props.editorRef(this._editor);
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
		return <div ref={(e) => this._container = e} {...otherProps} />;
	}
}
