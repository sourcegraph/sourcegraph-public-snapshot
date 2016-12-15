import "sourcegraph/editor/contrib";
import "sourcegraph/editor/FindExternalReferencesAction";
import "sourcegraph/editor/GotoDefinitionWithClickEditorContribution";
import "sourcegraph/editor/vscode";
import "sourcegraph/workbench/overrides/instantiationService";

import "vs/editor/common/editorCommon";
import "vs/editor/contrib/codelens/browser/codelens";
import "vs/workbench/browser/parts/editor/stringEditor";
import "vs/workbench/parts/files/browser/explorerViewlet";
import "vs/workbench/parts/files/browser/files.contribution";

import URI from "vs/base/common/uri";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { IOptions } from "vs/workbench/common/options";
import { Workbench } from "vs/workbench/electron-browser/workbench";

import { configureEditor } from "sourcegraph/editor/config";
import { configurePostStartup } from "sourcegraph/workbench/config";
import { setupServices } from "sourcegraph/workbench/setupServices";

// init creates the editor interface.
export function init(domElement: HTMLDivElement, resource: URI): [Workbench, ServiceCollection] {
	const workspace = resource.with({ fragment: "" });
	const services = setupServices(domElement, resource);
	const instantiationService = services.get(IInstantiationService) as IInstantiationService;

	const parent = domElement.parentElement;
	const workbench = instantiationService.createInstance(
		Workbench,
		parent,
		domElement,
		{ resource: workspace },
		options(resource),
		services,
	);
	workbench.startup();
	workbench.layout();

	const editor = workbench.getEditorPart();
	configureEditor(editor, resource);
	configurePostStartup(services);
	return [workbench, services];
}

function options(resource: URI): IOptions {
	return {
		filesToOpen: [
			{ resource },
		],
	};
}
