import * as React from "react";
import * as ReactDOM from "react-dom";

import { $ } from "vs/base/browser/builder";

import { ReferenceCard } from "sourcegraph/components";
import { OneReference } from "sourcegraph/workbench/info/referencesModel";

export function ReferenceItem(
	element: OneReference,
	container: HTMLElement,
	firstToggleAdded: boolean,
	refBaseHeight: number,
	refWithCommitInfoHeight: number,
): void {
	const preview = element.preview.preview(element.range);
	const fileName = element.uri.fragment.split("/").pop()!;
	const line = element.range.startLineNumber;
	const fnSignature = preview.before.concat(preview.inside, preview.after);
	const refContainer = $("div");
	let height = refBaseHeight;
	let defaultAvatar;
	let gravatarHash;
	let avatar;
	let authorName;
	let date;

	if (element && element.commitInfo && element.commitInfo.hunk.author && element.commitInfo.hunk.author.person) {
		defaultAvatar = "https://secure.gravatar.com/avatar?d=mm&f=y&s=128";
		gravatarHash = element.commitInfo.hunk.author.person.gravatarHash;
		avatar = gravatarHash ? `https://secure.gravatar.com/avatar/${gravatarHash}?s=128&d=retro` : defaultAvatar;
		authorName = element.commitInfo.hunk.author.person.name;
		date = element.commitInfo.hunk.author.date;
		height = refWithCommitInfoHeight;
	}

	refContainer.appendTo(container);

	ReactDOM.render(
		<ReferenceCard
			fnSignature={fnSignature}
			authorName={authorName}
			avatar={avatar}
			date={date}
			fileName={fileName}
			height={height}
			line={line} />,
		refContainer.getHTMLElement(),
	);
}
