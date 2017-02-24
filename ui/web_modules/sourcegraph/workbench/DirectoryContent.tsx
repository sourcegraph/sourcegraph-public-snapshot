import * as React from "react";
import * as ReactDOM from "react-dom";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";

import { FlexContainer } from "sourcegraph/components";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { Services } from "sourcegraph/workbench/services";

/**
 * renderDirectoryContent displays a welcome message when a user is viewing the root of or a directory in a repo.
 */
export function renderDirectoryContent(): void {
	const keyboardShortcutStyle = {
		backgroundColor: colors.blueGrayD1(),
		borderRadius: "3px",
		padding: "2px 5px",
	};
	const message = <div>
		Start by going to a file or hit <span style={keyboardShortcutStyle}>/</span > to search.
	</div>;
	renderRootContent(message);
}

/**
 * renderDirectoryContent displays an error message when a user navigates to a file that doesn't exist.
 */
export function renderNotFoundError(): void {
	const message = <div>
		404: File not found
	</div>;
	renderRootContent(message);
}

/**
 * renderRootContent displays a screen where the editor normally goes.
 */
function renderRootContent(message: JSX.Element): void {
	// We don't need or want the editor to be open when displaying the content for a directory.
	const editorService = Services.get(IWorkbenchEditorService) as IWorkbenchEditorService;
	editorService.closeAllEditors();

	const el = document.getElementById("workbench.parts.editor");
	if (!el) {
		throw new Error("Expected workbench.parts.editor element to exist.");
	}
	const container = el.firstChild;
	if (!container) {
		throw new Error("Expected workbench.parts.editor to have a child.");
	}

	const node = document.createElement("div");
	node.style.width = "100%";
	node.style.height = "100%";
	container.appendChild(node);

	const style = {
		fontFamily: typography.fontStack.sansSerif,
		color: colors.white(),
		margin: whitespace[2],
		textAlign: "center",
	};

	const content = <FlexContainer direction="top_bottom" justify="center" items="center" style={{
		width: "100%",
		height: "100%",
		padding: whitespace[2],
		paddingBottom: whitespace[6],
	}}>
		<div id="directory_help_message" style={{ ...style, ...typography.size[4] }}>{message}</div>
	</FlexContainer>;
	ReactDOM.render(content, node);
}
