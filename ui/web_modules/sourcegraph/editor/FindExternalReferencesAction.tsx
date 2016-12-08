import { TPromise } from "vs/base/common/winjs.base";
import { ICommonCodeEditor, IPosition, IReadOnlyModel, ModeContextKeys } from "vs/editor/common/editorCommon";
import { EditorAction, ServicesAccessor, editorAction } from "vs/editor/common/editorCommonExtensions";
import { Location } from "vs/editor/common/modes";
import { getDeclarationsAtPosition } from "vs/editor/contrib/goToDeclaration/common/goToDeclaration";
import { ReferencesController } from "vs/editor/contrib/referenceSearch/browser/referencesController";
import { ReferencesModel } from "vs/editor/contrib/referenceSearch/browser/referencesModel";
import { ContextKeyExpr } from "vs/platform/contextkey/common/contextkey";

import * as BlobActions from "sourcegraph/blob/BlobActions";
import { URIUtils } from "sourcegraph/core/uri";
import * as Dispatcher from "sourcegraph/Dispatcher";
import { makeRepoRev } from "sourcegraph/repo";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { Features } from "sourcegraph/util/features";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

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

	private _findExternalReferences(model: IReadOnlyModel, pos: IPosition): TPromise<void> {
		const {repo, rev, path} = URIUtils.repoParams(model.uri);
		AnalyticsConstants.Events.CodeExternalReferences_Viewed.logEvent({ repo, rev: rev || "", path });

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

@editorAction
class PeekExternalReferences extends EditorAction {
	constructor() {
		if (Features.externalReferences.isEnabled()) {
			super({
				id: "peek.external.references",
				label: "Peek External References",
				alias: "Peek External References",
				precondition: ContextKeyExpr.and(ModeContextKeys.hasReferenceProvider),
				menuOpts: {
					group: "navigation",
					order: 1.3,
				},
			});
		}
	}

	run(accessor: ServicesAccessor, editor: ICommonCodeEditor): void {
		getDeclarationsAtPosition(editor.getModel(), editor.getPosition()).then(references => {
			let result: Location[] = [];
			// Example of how to add to the peek view.
			// result.push({
			// 	range: {
			// 		startColumn: 26,
			// 		startLineNumber: 233,
			// 		endLineNumber: 233,
			// 		endColumn: 40,
			// 	},
			// 	uri: URI.from({
			// 		scheme: "git",
			// 		authority: "github.com",
			// 		fragment: "src/time/time.go",
			// 		path: "/golang/go",
			// 		query: "0d818588685976407c81c60d2fda289361cbc8ec",
			// 	}),
			// });
			const controller = ReferencesController.get(editor);
			controller.toggleWidget(editor.getSelection(), TPromise.as(new ReferencesModel(result)), {
				getMetaTitle: () => {
					return "(Placeholder) External References";
				},
				onGoto: () => {
					controller.closeWidget();
					return TPromise.as(editor);
				},
			});
		});
	}

}
