import { TPromise } from "vs/base/common/winjs.base";
import { ICommonCodeEditor, IPosition, IReadOnlyModel, ModeContextKeys } from "vs/editor/common/editorCommon";
import { EditorAction, ServicesAccessor, editorAction } from "vs/editor/common/editorCommonExtensions";
import { ContextKeyExpr } from "vs/platform/contextkey/common/contextkey";

import * as BlobActions from "sourcegraph/blob/BlobActions";
import { URIUtils } from "sourcegraph/core/uri";
import * as Dispatcher from "sourcegraph/Dispatcher";
import { makeRepoRev } from "sourcegraph/repo";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

const fetch = singleflightFetch(defaultFetch);

@editorAction
export class FindExternalReferencesAction extends EditorAction {
	constructor() {
		super({
			id: "editor.action.findExternalReferences",
			label: "Find External References",
			alias: "Find External References",
			precondition: ContextKeyExpr.and(ModeContextKeys.hasReferenceProvider),
			menuOpts: {
				group: "navigation",
				order: 1.4,
			},
		});
	}

	public run(accessor: ServicesAccessor, editor: ICommonCodeEditor): void {
		this.findExternalReferences(editor.getModel(), editor.getPosition());
	}

	private findExternalReferences(model: IReadOnlyModel, pos: IPosition): TPromise<void> {
		const { repo, rev, path } = URIUtils.repoParams(model.uri);
		Events.CodeExternalReferences_Viewed.logEvent({ repo, rev: rev || "", path });

		const line = pos.lineNumber - 1;
		const col = pos.column - 1;
		return TPromise.wrap<void>(fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/def-landing?file=${path}&line=${line}&character=${col}`)
			.then(checkStatus)
			.then(resp => resp.json())
			.catch(err => null)
			.then((resp) => {
				if (resp && (resp as any).URL) {
					window.location.href = (resp as any).URL;
				} else {
					Dispatcher.Stores.dispatch(new BlobActions.Toast("No external references found"));
				}
			}));
	}
}
