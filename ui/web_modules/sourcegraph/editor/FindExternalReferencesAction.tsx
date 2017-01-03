import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { ICommonCodeEditor, IPosition, IReadOnlyModel, ModeContextKeys } from "vs/editor/common/editorCommon";
import { EditorAction, ServicesAccessor, editorAction } from "vs/editor/common/editorCommonExtensions";
import { Location } from "vs/editor/common/modes";
import { ReferencesController } from "vs/editor/contrib/referenceSearch/browser/referencesController";
import { ReferencesModel } from "vs/editor/contrib/referenceSearch/browser/referencesModel";
import { PeekContext } from "vs/editor/contrib/zoneWidget/browser/peekViewWidget";
import { ContextKeyExpr } from "vs/platform/contextkey/common/contextkey";

import * as BlobActions from "sourcegraph/blob/BlobActions";
import { URIUtils } from "sourcegraph/core/uri";
import * as Dispatcher from "sourcegraph/Dispatcher";
import { makeRepoRev } from "sourcegraph/repo";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { Features } from "sourcegraph/util/features";
import { RefData, resolveGlobalReferences } from "sourcegraph/util/GlobalRefsBackend";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

const fetch = singleflightFetch(defaultFetch);

@editorAction
class FindExternalReferencesAction extends EditorAction {
	constructor() {
		if (!Features.externalReferences.isEnabled()) {
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
				precondition: ContextKeyExpr.and(ModeContextKeys.hasReferenceProvider, PeekContext.notInPeekEditor),
				menuOpts: {
					group: "navigation",
					order: 1.3,
				},
			});
		}
	}

	run(accessor: ServicesAccessor, editor: ICommonCodeEditor): void {
		return this._getGlobalReferencesAtPosition(editor);
	}

	_getGlobalReferencesAtPosition(editor: ICommonCodeEditor): void {
		const controller = ReferencesController.get(editor);
		const { repo, path } = URIUtils.repoParams(editor.getModel().uri);
		const editorPosition = editor.getPosition();
		let model = editor.getModel();
		let refData: RefData = {
			language: model.getModeIdAtPosition(editorPosition.lineNumber, editorPosition.column),
			repo: repo,
			version: model.uri.query,
			file: path,
			line: editorPosition.lineNumber - 1,
			column: editorPosition.column - 1,
		};

		resolveGlobalReferences(refData).then((globalRefs) => {
			let globalRefLocs: Location[] = [];
			globalRefs.forEach((ref) => {
				if (!ref.refLocation || !ref.uri) {
					return;
				}
				globalRefLocs.push({
					range: {
						startLineNumber: ref.refLocation.startLineNumber + 1,
						startColumn: ref.refLocation.startColumn + 1,
						endLineNumber: ref.refLocation.endLineNumber + 1,
						endColumn: ref.refLocation.endColumn + 1,
					},
					uri: URI.from({
						scheme: ref.uri.scheme,
						query: ref.uri.query,
						path: ref.uri.path,
						fragment: ref.uri.fragment,
						authority: ref.uri.host,
					}),
				});
			});

			const refModel = TPromise.as(new ReferencesModel(globalRefLocs));
			controller.toggleWidget(editor.getSelection(), refModel, {
				getMetaTitle: () => {
					return "";
				},
				onGoto: () => {
					controller.closeWidget();
					return TPromise.as(editor);
				},
			});
		})
			.catch(err => err);
	}
}

// Suppress tsc noUnusedLocals errors. These are used via the @editorAction decorator.
FindExternalReferencesAction; // tslint:disable-line no-unused-expression
PeekExternalReferences; // tslint:disable-line no-unused-expression
