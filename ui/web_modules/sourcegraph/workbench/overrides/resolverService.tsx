import { always, first } from "vs/base/common/async";
import { IDisposable, IReference, ImmortalReference, ReferenceCollection, toDisposable } from "vs/base/common/lifecycle";
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

class ResourceModelCollection extends ReferenceCollection<TPromise<ITextEditorModel>> {

	private providers: { [scheme: string]: ITextModelContentProvider[] } = Object.create(null);

	constructor(
		@IInstantiationService private instantiationService: IInstantiationService
	) {
		super();
	}

	createReferencedObject(key: string): TPromise<ITextEditorModel> {
		const resource = URI.parse(key);

		return this.resolveTextModelContent(key)
			.then(() => this.instantiationService.createInstance(ResourceEditorModel, resource));
	}

	destroyReferencedObject(modelPromise: TPromise<ITextEditorModel>): void {
		modelPromise.done(model => model.dispose());
	}

	// VSCode implementation - TODO: Remove if overrides are not needed. - @Kingy
	registerTextModelContentProvider(scheme: string, provider: ITextModelContentProvider): IDisposable {
		const registry = this.providers;
		const providers = registry[scheme] || (registry[scheme] = []);

		providers.unshift(provider);

		return toDisposable(() => {
			const array = registry[scheme];

			if (!array) {
				return;
			}
			const index = array.indexOf(provider);

			if (index === -1) {
				return;
			}

			array.splice(index, 1);

			if (array.length === 0) {
				delete registry[scheme];
			}
		});
	}

	private resolveTextModelContent(key: string): TPromise<IModel> {
		const resource = URI.parse(key);
		const providers = this.providers[resource.scheme] || [];
		const factories = providers.map(p => () => p.provideTextContent(resource));

		return first(factories).then(model => {
			if (!model) {
				console.error(`Could not resolve any model with uri '${resource}'.`);
				return TPromise.wrapError("Could not resolve any model with provided uri.");
			}

			return model;
		});
	}
}

export class TextModelResolverService implements ITextModelResolverService {
	public _serviceBrand: any;

	private contentProvider: ITextModelContentProvider;
	private resourceModelCollection: ResourceModelCollection;
	private modelService: IModelService;

	private promiseCache: { [uri: string]: TPromise<IReference<ITextEditorModel>> } = Object.create(null);

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
		this.resourceModelCollection = instantiationService.createInstance(ResourceModelCollection);
		this.modelService = modelService;
	}

	createModelReference(resource: URI): TPromise<IReference<ITextEditorModel>> {
		const uri = resource.toString();
		let promise = this.promiseCache[uri];

		if (promise) {
			return promise;
		}

		promise = this.promiseCache[uri] = this._createModelReference(resource);

		return always(promise, () => delete this.promiseCache[uri]);
	}

	private _createModelReference(resource: URI): TPromise<IReference<ITextEditorModel>> {
		// resource.scheme === "zap" - Takes a Zap op.create and turns it into a new file with a standard uri.
		// there is no special casing with zap other than on creating and deleting files/directories that do not exist at the git base commit level.
		// URIs that are loaded and passed on Sourcegraph will never contain "zap" in the scheme unless they are the result
		// of an OT op and not currently committed to Git, but they are resolved and converted to "Git" here before any of their usage with a zap scheme is used elsewhere.
		// This can be improved, but I would not consider it a hack since the underlying call relies on `openTextDocument` to create a new doc.
		// We could entertain the idea of using "untitled" to create new virtual documents instead of "zap" later,
		// but passing a new scheme and registring it in the way we do follows the standard for how a TextDocumentContentProvider is created and used.
		if (resource.scheme === "git" && URIUtils.hasAbsoluteCommitID(resource) || resource.scheme === "zap") {
			return this.textFileService.models.loadOrCreate(resource).then(model => {
				return this.modeService.getOrCreateModeByFilenameOrFirstLine(resource.fragment).then(mode => {
					model.textEditorModel.setMode(mode.getId());
					return new ImmortalReference(model);
				});
			});
		}

		// InMemory Schema: go through model service cache
		// TODO ImmortalReference is a hack
		if (resource.scheme === "inmemory") {
			const cachedModel = this.modelService.getModel(resource);

			if (!cachedModel) {
				return TPromise.wrapError("Cant resolve inmemory resource");
			}

			return TPromise.as(new ImmortalReference(this.instantiationService.createInstance(ResourceEditorModel, resource)));
		}

		const ref = this.resourceModelCollection.acquire(resource.toString());
		return ref.object.then(model => ({ object: model, dispose: () => ref.dispose() }));
	}

	registerTextModelContentProvider(scheme: string, provider: ITextModelContentProvider): IDisposable {
		return this.resourceModelCollection.registerTextModelContentProvider(scheme, provider);
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
