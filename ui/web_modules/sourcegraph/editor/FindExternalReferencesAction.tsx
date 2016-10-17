import * as BlobActions from "sourcegraph/blob/BlobActions";
import {URIUtils} from "sourcegraph/core/uri";
import {urlToDefInfo} from "sourcegraph/def/routes";

import * as Dispatcher from "sourcegraph/Dispatcher";
import { makeRepoRev } from "sourcegraph/repo";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

import {TPromise} from "vs/base/common/winjs.base";
import {ICommonCodeEditor, IPosition, IReadOnlyModel, ModeContextKeys} from "vs/editor/common/editorCommon";
import {EditorAction, ServicesAccessor, editorAction} from "vs/editor/common/editorCommonExtensions";
import {ContextKeyExpr} from "vs/platform/contextkey/common/contextkey";

const fetch = singleflightFetch(defaultFetch);

@editorAction
class FindExternalReferencesAction extends EditorAction {
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
		this._findExternalReferences(editor.getModel(), editor.getPosition());
	}

	private	_findExternalReferences(model: IReadOnlyModel, pos: IPosition): TPromise<void> {
		const {repo, rev, path} = URIUtils.repoParams(model.uri);
		EventLogger.logEventForCategory(
			AnalyticsConstants.CATEGORY_REFERENCES,
			AnalyticsConstants.ACTION_CLICK,
			"ClickedViewExternalReferences",
			{ repo, rev: rev || "", path }
		);

		const line = pos.lineNumber - 1;
		const col = pos.column - 1;
		return TPromise.wrap<void>(fetch(`/.api/repos/${makeRepoRev(repo, rev)}/-/hover-info?file=${path}&line=${line}&character=${col}`)
			.then(checkStatus)
			.then(resp => resp.json())
			.catch(err => null)
			.then((resp) => {
				if (resp && (resp as any).def) {
					// TODO(uforic): Remove this when we remove srclib dependency. Fix a special case for golang/go. 
					const def = resp.def;
					if (def.Repo === "github.com/golang/go" && def.Unit && def.Unit.startsWith("github.com/golang/go/src/")) {
						def.Unit = def.Unit.replace("github.com/golang/go/src/", "");
					}
					window.location.href = urlToDefInfo((resp as any).def);
				} else {
					Dispatcher.Stores.dispatch(new BlobActions.Toast("No external references found"));
				}
			}));
	}
}
