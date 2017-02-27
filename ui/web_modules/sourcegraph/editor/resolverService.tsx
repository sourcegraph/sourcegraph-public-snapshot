import { IDisposable, IReference, ImmortalReference } from "vs/base/common/lifecycle";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IModel } from "vs/editor/common/editorCommon";
import { IModelService } from "vs/editor/common/services/modelService";
import { IModeService } from "vs/editor/common/services/modeService";
import { ITextEditorModel, ITextModelContentProvider, ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { ResourceEditorModel } from "vs/workbench/common/editor/resourceEditorModel";
import { ITextFileService } from "vs/workbench/services/textfile/common/textfiles";

import { URIUtils } from "sourcegraph/core/uri";

export class TextModelResolverService implements ITextModelResolverService {
	public _serviceBrand: any;

	private contentProvider: ITextModelContentProvider;

	constructor(
		@IModelService modelService: IModelService,
		@IModeService private modeService: IModeService,
		@ITextFileService private textFileService: ITextFileService,
		@IInstantiationService private instantiationService: IInstantiationService,
	) {
		this.contentProvider = new TextModelContentProvider(
			modeService,
			textFileService,
		);
	}

	createModelReference(resource: URI): TPromise<IReference<ITextEditorModel>> {
		if (resource.scheme === "git" && URIUtils.hasAbsoluteCommitID(resource)) {
			return this.textFileService.models.loadOrCreate(resource).then(model => {
				return this.modeService.getOrCreateModeByFilenameOrFirstLine(resource.fragment).then(mode => {
					model.textEditorModel.setMode(mode.getId());
					return new ImmortalReference(model);
				});
			});
		}
		return this.contentProvider.provideTextContent(resource).then((model) =>
			new ImmortalReference(this.instantiationService.createInstance(ResourceEditorModel, resource)),
		);
	}

	registerTextModelContentProvider(scheme: string, provider: ITextModelContentProvider): IDisposable {
		return {
			dispose: () => { /* */ },
		};
	}

}

export class TextModelContentProvider implements ITextModelContentProvider {

	constructor(
		@IModeService private modeService: IModeService,
		@ITextFileService private textFileService: ITextFileService,
	) {
		//
	}

	provideTextContent(resource: URI): TPromise<IModel> {
		return this.textFileService.models.loadOrCreate(resource).then(model => {
			return this.modeService.getOrCreateModeByFilenameOrFirstLine(resource.fragment).then(mode => {
				model.textEditorModel.setMode(mode.getId());
				return model.textEditorModel;
			});
		});
	}
}
