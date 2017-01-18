import { TPromise } from "vs/base/common/winjs.base";
import { ICommonCodeEditor, IPosition, IReadOnlyModel, ModeContextKeys } from "vs/editor/common/editorCommon";
import { EditorAction, ServicesAccessor, editorAction } from "vs/editor/common/editorCommonExtensions";
import { PeekContext } from "vs/editor/contrib/zoneWidget/browser/peekViewWidget";
import { ContextKeyExpr } from "vs/platform/contextkey/common/contextkey";
import { IEditorService } from "vs/platform/editor/common/editor";

import * as BlobActions from "sourcegraph/blob/BlobActions";
import { URIUtils } from "sourcegraph/core/uri";
import * as Dispatcher from "sourcegraph/Dispatcher";
import { makeRepoRev } from "sourcegraph/repo";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { Features } from "sourcegraph/util/features";
import { provideGlobalReferences } from "sourcegraph/util/RefsBackend";
import { singleflightFetch } from "sourcegraph/util/singleflightFetch";
import { checkStatus, defaultFetch } from "sourcegraph/util/xhr";
import { ReferencesController } from "vs/editor/contrib/referenceSearch/browser/referencesController";

const fetch = singleflightFetch(defaultFetch);

@editorAction
export class FindExternalReferencesAction extends EditorAction {
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
		this.findExternalReferences(editor.getModel(), editor.getPosition());
	}

	private findExternalReferences(model: IReadOnlyModel, pos: IPosition): TPromise<void> {
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
export class PeekExternalReferences extends EditorAction {
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
		const editorService = accessor.get(IEditorService);
		const refModel = provideGlobalReferences(editor);
		const controller = ReferencesController.get(editor);
		controller.toggleWidget(editor.getSelection(), refModel as any, {
			getMetaTitle: () => {
				return "";
			},
			onGoto: (ref) => {
				controller.closeWidget();
				return editorService.openEditor({
					resource: ref.uri,
					options: {
						selection: ref.range
					}
				});
			},
		});
	}
}
