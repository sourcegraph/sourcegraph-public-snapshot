import { EditorService } from "sourcegraph/editor/EditorService";
import { IEditorConstructionOptions, IStandaloneCodeEditor, StandaloneEditor } from "vs/editor/browser/standalone/standaloneCodeEditor";
import { DynamicStandaloneServices } from "vs/editor/browser/standalone/standaloneServices";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { ICommandService } from "vs/platform/commands/common/commands";
import { IEditorService } from "vs/platform/editor/common/editor";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { OpenerService } from "vs/platform/opener/browser/openerService";
import { IOpenerService } from "vs/platform/opener/common/opener";
import { IUntitledEditorService, UntitledEditorService } from "vs/workbench/services/untitled/common/untitledEditorService";

import { TextModelResolverService } from "sourcegraph/editor/resolverService";

export function createEditor(domElement: HTMLElement, options: IEditorConstructionOptions): [IStandaloneCodeEditor, EditorService] {
	let services = new DynamicStandaloneServices(domElement, {});

	const editorService = new EditorService();
	services.set(IEditorService, editorService);

	const instantiationService = services.get(IInstantiationService);
	services.set(IUntitledEditorService, instantiationService.createInstance(UntitledEditorService));
	services.set(ITextModelResolverService, instantiationService.createInstance(TextModelResolverService));
	services.set(IOpenerService, new OpenerService(services.get(IEditorService), services.get(ICommandService)));

	const editor = instantiationService.createInstance(
		StandaloneEditor,
		domElement,
		options,
		services,
	);

	editorService.setEditor(editor);

	return [editor, editorService];
}
