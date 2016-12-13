import "sourcegraph/editor/contrib";
import "sourcegraph/editor/FindExternalReferencesAction";
import "sourcegraph/editor/GotoDefinitionWithClickEditorContribution";
import "sourcegraph/editor/vscode";
import "sourcegraph/workbench/overrides/fileService";
import "sourcegraph/workbench/overrides/iconLabel";

import "vs/editor/common/editorCommon";
import "vs/editor/contrib/codelens/browser/codelens";
import "vs/workbench/parts/files/browser/files.contribution";

import URI from "vs/base/common/uri";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { Workbench } from "vs/workbench/electron-browser/workbench";

import { setConfiguration } from "sourcegraph/workbench/config";
import { setupServices } from "sourcegraph/workbench/setupServices";

export function init(domElement: HTMLDivElement): Workbench {
	const services = setupServices(domElement);
	const instantiationService = services.get(IInstantiationService) as IInstantiationService;
	setConfiguration(services);

	const parent = domElement.parentElement;
	const workbench = instantiationService.createInstance(
		Workbench,
		parent,
		domElement,
		workspace,
		options,
		services,
	);
	workbench.startup();
	workbench.layout();
	return workbench;
}

const options = {};

const workspace = {
	resource: URI.parse("foo://github.com/sourcegraph/sourcegraph"),
};
