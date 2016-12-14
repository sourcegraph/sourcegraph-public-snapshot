import "sourcegraph/editor/contrib";
import "sourcegraph/editor/FindExternalReferencesAction";
import "sourcegraph/editor/GotoDefinitionWithClickEditorContribution";
import "sourcegraph/editor/vscode";
import "sourcegraph/workbench/overrides/fileService";
import "sourcegraph/workbench/overrides/labels";

import "vs/editor/common/editorCommon";
import "vs/editor/contrib/codelens/browser/codelens";
import "vs/workbench/browser/parts/editor/stringEditor";
import "vs/workbench/parts/files/browser/explorerViewlet";
import "vs/workbench/parts/files/browser/files.contribution";

import URI from "vs/base/common/uri";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { Workbench } from "vs/workbench/electron-browser/workbench";

import { configureEditor } from "sourcegraph/editor/config";
import { configureServices, configurePostStartup } from "sourcegraph/workbench/config";
import { setupServices } from "sourcegraph/workbench/setupServices";

export function init(domElement: HTMLDivElement, resource: URI): Workbench {
	const workspace = resource.with({fragment: ""});
	const services = setupServices(domElement);
	const instantiationService = services.get(IInstantiationService) as IInstantiationService;
	configureServices(services, workspace);

	const parent = domElement.parentElement;
	const workbench = instantiationService.createInstance(
		Workbench,
		parent,
		domElement,
		{resource: workspace},
		options,
		services,
	);
	workbench.startup();
	workbench.layout();

	const editor = workbench.getEditorPart();
	configureEditor(editor, resource, instantiationService);
	configurePostStartup(services);
	return workbench;
}

const options = {};
