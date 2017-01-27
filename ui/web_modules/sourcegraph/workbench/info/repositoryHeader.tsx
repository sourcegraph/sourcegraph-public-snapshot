import { css, insertGlobal } from "glamor";
import * as React from "react";

import { $, Builder } from "vs/base/browser/builder";
import { IDisposable } from "vs/base/common/lifecycle";
import URI from "vs/base/common/uri";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";

import { FileReferences } from "sourcegraph/workbench/info/referencesModel";
import { WorkspaceBadge } from "sourcegraph/workbench/ui/badges/workspaceBadge";
import { FileLabel } from "sourcegraph/workbench/ui/fileLabel";
import { LeftRightWidget } from "sourcegraph/workbench/ui/leftRightWidget";

import { List } from "sourcegraph/components/symbols/Primaries";
import { colors, paddingMargin, whitespace } from "sourcegraph/components/utils";

export function RepositoryHeader(
	element: FileReferences,
	container: HTMLElement,
	userToggledState: Set<string>,
	firstToggleAdded: boolean,
	fileRefsHeight: number,
	contextService: IWorkspaceContextService,
	editorUriPath: string,
): void {
	const repositoryHeader = $(".refs-repository-title");

	const borderSx = `1px solid ${colors.blueGrayL1(0.2)}`;
	const refHeaderSx = css(
		paddingMargin.padding("x", 2),
		{
			backgroundColor: colors.blueGrayL3(),
			borderBottom: borderSx,
			borderTop: borderSx,
			boxShadow: `0 2px 2px 0 ${colors.black(0.05)}`,
			color: colors.text(),
			display: "flex",
			fontWeight: "bold",
			alignItems: "center",
			height: fileRefsHeight,
		},
	).toString();

	const rowSx = css({
		paddingLeft: whitespace[2],
		paddingRight: whitespace[2],
	}).toString();

	let workspaceURI = URI.from({
		scheme: element.uri.scheme,
		authority: element.uri.authority,
		path: element.uri.path,
		query: element.uri.query,
		fragment: element.uri.path,
	});

	// tslint:disable:no-new
	new LeftRightWidget(repositoryHeader, left => {
		const repoTitleContent = new FileLabel(left, workspaceURI, contextService);
		repoTitleContent.setIcon(<List width={18} style={{ marginLeft: -2, color: colors.blueGrayL1() }} />);
		return null as any;
	}, right => {

		const workspace = workspaceURI.path === editorUriPath ? "Local" : "External";
		const badge = new WorkspaceBadge(right, workspace);

		if (element.failure) {
			badge.setTitleFormat("Failed to resolve file.");
		} else if (workspace === "Local") {
			badge.setTitleFormat("Local");
			badge.setColor(colors.green());
		} else {
			badge.setTitleFormat("External");
			badge.setColor(colors.orangeL1());
		}

		return badge as any;
	}).setClassNames(rowSx);

	setStyles(refHeaderSx);
	repositoryHeader.getHTMLElement().classList.add(refHeaderSx);
	repositoryHeader.on("click", (e: Event, builder: Builder, unbind: IDisposable): void => {
		const stateKey = element.uri.toString();
		if (userToggledState.has(stateKey)) {
			userToggledState.delete(stateKey);
		} else {
			userToggledState.add(stateKey);
		}
	});
	repositoryHeader.appendTo(container);
}

function setStyles(className: string): void {

	insertGlobal(`.monaco-tree-row.has-children .${className}:before`, {
		content: `""`,
		height: 15,
		width: 9,
		marginLeft: 9,
		transition: "all 300ms ease-in-out",
		backgroundRepeat: "no-repeat",
		backgroundImage: "url(data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNyIgaGVpZ2h0PSIxMiIgdmlld0JveD0iMjQgMTUgNyAxMiIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNMjUuNDcyIDI2LjE0NWMtLjI2LjI2LS42ODIuMjYtLjk0MyAwLS4yNjItLjI2LS4yNjItLjY4Mi0uMDAyLS45NDNsNC41MzYtNC41My00LjUzNi00LjUzYy0uMjYtLjI2LS4yNi0uNjgzIDAtLjk0My4yNjItLjI2Mi42ODQtLjI2Ljk0NCAwbDUuMDA4IDVjLjEyNS4xMjYuMTk2LjI5NS4xOTYuNDcycy0uMDcuMzQ3LS4xOTYuNDcybC01LjAwOCA1eiIgZmlsbD0iIzc3OTNBRSIgZmlsbC1ydWxlPSJldmVub2RkIi8+PC9zdmc+)",
	});

	insertGlobal(`.monaco-tree-row.has-children.expanded .${className}:before`, {
		transform: "rotate(90deg)",
	});

	insertGlobal(`.monaco-tree-row.has-children .${className} .label-name`, {
		width: 159,
		verticalAlign: "top",
		textAlign: "left",
		display: "inline-block",
		overflow: "hidden",
		textOverflow: "ellipsis",
		direction: "rtl",
	});
}
