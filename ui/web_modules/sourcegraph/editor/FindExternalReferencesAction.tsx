import URI from "vs/base/common/uri";
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
import { RefData, resolveGlobalReferences } from "sourcegraph/util/GlobalRefsBackend";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";

const fetch = singleflightFetch(defaultFetch);

function cacheKey(model: IReadOnlyModel, position: IPosition): string {
	return `${model.uri.toString(true)}:${position.lineNumber}:${position.column}`;
}

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

const globalRefsCache = new Map<string, TPromise<ReferencesModel>>();

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
		return this._getGlobalReferencesAtPosition(editor);
	}

	_getGlobalReferencesAtPosition(editor: ICommonCodeEditor): void {
		getDeclarationsAtPosition(editor.getModel(), editor.getPosition()).then(declLocs => {
			if (!declLocs || declLocs.length === 0) {
				return TPromise.as(null);
			}

			const loc = declLocs[0];
			const key = cacheKey(editor.getModel(), {
				lineNumber: loc.range.startLineNumber,
				column: loc.range.startColumn,
			});
			const cached = globalRefsCache.get(key);
			const controller = ReferencesController.get(editor);

			if (cached) {
				return controller.toggleWidget(editor.getSelection(), cached, {
					getMetaTitle: () => {
						return "";
					},
					onGoto: () => {
						controller.closeWidget();
						return TPromise.as(editor);
					},
				});
			}

			const { repo, path } = URIUtils.repoParams(editor.getModel().uri);
			let uri = editor.getModel().uri;
			let refData: RefData = {
				repo: repo,
				version: uri.query,
				file: path,
				line: loc.range.startLineNumber,
				column: loc.range.startColumn,
			};

			resolveGlobalReferences(refData).then((globalRefs) => {
				let globalRefLocs: Location[] = [];
				globalRefs.forEach((ref) => {
					if (!ref.refLocation || !ref.uri) {
						return;
					}
					globalRefLocs.push({
						range: {
							startLineNumber: ref.refLocation.startLineNumber,
							startColumn: ref.refLocation.startColumn,
							endLineNumber: ref.refLocation.endLineNumber,
							endColumn: ref.refLocation.endColumn,
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
				globalRefsCache.set(key, refModel);
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
		});
	}
}
