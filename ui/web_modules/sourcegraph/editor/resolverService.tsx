import { IDisposable, IReference, ImmortalReference } from "vs/base/common/lifecycle";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IModel } from "vs/editor/common/editorCommon";
import { IModelService } from "vs/editor/common/services/modelService";
import { IModeService } from "vs/editor/common/services/modeService";
import { ITextEditorModel, ITextModelContentProvider, ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { EditorModel } from "vs/workbench/common/editor";

import { fetchContent } from "sourcegraph/editor/contentLoader";

export class TextModelResolverService implements ITextModelResolverService {
	public _serviceBrand: any;

	private contentProvider: ITextModelContentProvider;

	constructor(
		@IModelService modelService: IModelService,
		@IModeService modeService: IModeService,
	) {
		this.contentProvider = new TextModelContentProvider(
			modelService,
			modeService,
		);
	}

	createModelReference(resource: URI): TPromise<IReference<ITextEditorModel>> {
		return this.contentProvider.provideTextContent(resource).then((model) =>
			new ImmortalReference(new TextEditorModel(model))
		);
	}

	registerTextModelContentProvider(scheme: string, provider: ITextModelContentProvider): IDisposable {
		return {
			dispose: () => { /* */ },
		};
	}

}

class TextEditorModel extends EditorModel {
	textEditorModel: IModel;

	constructor(model: IModel) {
		super();
		this.textEditorModel = model;
	}
}

export class TextModelContentProvider implements ITextModelContentProvider {

	constructor(
		@IModelService private modelService: IModelService,
		@IModeService private modeService: IModeService,
	) {
		//
	}

	provideTextContent(resource: URI): TPromise<IModel> {
		let model = this.modelService.getModel(resource);
		if (model) {
			return TPromise.wrap(model);
		}
		return fetchContent(resource).then((content) => {
			model = this.modelService.getModel(resource);
			if (model) {
				return model;
			}
			const mode = this.modeService.getOrCreateModeByFilenameOrFirstLine(resource.fragment);
			return this.modelService.createModel(content, mode, resource);
		});
	}
}
