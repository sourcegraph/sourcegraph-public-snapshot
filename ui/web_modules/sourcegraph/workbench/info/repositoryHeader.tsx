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

import { ChevronDown, List } from "sourcegraph/components/symbols/Primaries";
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
		{
			":hover": {
				borderColor: colors.blueGrayL2(),
				color: colors.blue(),
			}
		}
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
	const repoHeaderWidget = new LeftRightWidget(repositoryHeader, left => {
		const repoTitleContent = new FileLabel(left, workspaceURI, contextService);
		repoTitleContent.setIcon(<List width={18} style={{ marginLeft: -2, opacity: 0.8 }} />);
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
	});

	repoHeaderWidget.setClassNames(rowSx);
	repoHeaderWidget.setIcon(<ChevronDown />);

	setGlobalStyles(refHeaderSx);

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

function setGlobalStyles(className: string): void {

	const disclosureIconSx = {
		default: { transition: "transform 200ms" },
		closed: { transform: "rotate(-90deg)" },
		open: { transform: "rotate(0deg)" },
	};

	const repoNameSx = {
		width: 150,
		verticalAlign: "top",
		textAlign: "left",
		display: "inline-block",
		overflow: "hidden",
		textOverflow: "ellipsis",
		direction: "rtl",
	};

	const selectedSx = {
		backgroundColor: colors.blue(),
		color: "white",
		boxShadow: `0 2px 7px ${colors.black(0.2)}`
	};

	insertGlobal(`.monaco-tree-row.has-children.expanded .left-right-widget_icon`, { ...disclosureIconSx.default, ...disclosureIconSx.open });
	insertGlobal(`.monaco-tree-row.has-children .left-right-widget_icon`, { ...disclosureIconSx.default, ...disclosureIconSx.closed });
	insertGlobal(`.monaco-tree-row.has-children.selected .${className}`, selectedSx);
	insertGlobal(`.monaco-tree-row.has-children .${className} .label-name`, repoNameSx);
}
