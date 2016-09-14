// tslint:disable: typedef ordered-imports

import { urlTo } from "sourcegraph/util/urlTo";
import { urlToTree } from "sourcegraph/tree/routes";
import { urlToRepoRev } from "sourcegraph/repo/routes";
import { Def } from "sourcegraph/api";

// urlToDefInfo returns a URL to the given def's info at the given revision.
export function urlToDefInfo(def: Def, rev?: string | null): string {
	if ((def.File === null || def.Kind === "package")) {
		// The def's File field refers to a directory (e.g., in the
		// case of a Go package). We can't show a dir in this view,
		// so just redirect to the dir listing.
		//
		// TODO(sqs): Improve handling of this case.
		let file = def.File === "." ? "" : def.File;
		return urlToTree(def.Repo || "", rev || null, file);
	}
	return `${urlToRepoRev(def.Repo || "", rev || "")}/-/info/${def.UnitType}/${def.Unit}/-/${def.Path}`;
}

// urlToRepoBlob returns a URL to the given repositories file at the given revision.
export function urlToRepoBlob(repo: string, rev: string | null, blob: string): string {
	const revPart = rev ? `@${rev}` : "";
	return urlTo("blob", { splat: [`${repo}${revPart}`, blob] } as any);
}
