import "sourcegraph/feedback/feedback.contribution";

import "sourcegraph/editor/authorshipCodeLens";
import "sourcegraph/editor/authorshipWidget";
import "sourcegraph/editor/vscode";
import "sourcegraph/workbench/info/contrib";
import "sourcegraph/workbench/staticImports";

import "vs/editor/common/editorCommon";
import "vs/editor/contrib/codelens/browser/codelens";
import "vs/workbench/browser/parts/editor/stringEditor";
import "vs/workbench/parts/files/browser/explorerViewlet";
import "vs/workbench/parts/files/browser/files.contribution";
import "vs/workbench/parts/output/browser/output.contribution";
import "vs/workbench/parts/quickopen/browser/quickopen.contribution";
import "vs/workbench/parts/search/browser/search.contribution";
import "vs/workbench/parts/search/browser/searchViewlet";

import { start as startTappingOutputChannel } from "sourcegraph/workbench/outputListener";
import { IDisposable } from "vs/base/common/lifecycle";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IMode } from "vs/editor/common/modes";
import { IModeService } from "vs/editor/common/services/modeService";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import "vs/workbench/electron-browser/main.contribution";
import { Workbench } from "vs/workbench/electron-browser/workbench";
import { TextFileEditorModel } from "vs/workbench/services/textfile/common/textFileEditorModel";
import { ITextFileService } from "vs/workbench/services/textfile/common/textfiles";

import { TextModelResolverService } from "sourcegraph/editor/resolverService";
import { init as initExtensionHost } from "sourcegraph/ext/main";
import { Features } from "sourcegraph/util/features";
import { configurePostStartup, configurePreStartup } from "sourcegraph/workbench/config";
import { setupServices } from "sourcegraph/workbench/services";
import { GitTextFileService } from "sourcegraph/workbench/textFileService";

/**
 * init bootraps workbench creation.
 */
export function init(domElement: HTMLDivElement, resource: URI): [Workbench, ServiceCollection] {
	const workspace = resource.with({ fragment: "" });
	const services = setupServices(domElement, workspace);
	configurePreStartup(services);

	const instantiationService = services.get(IInstantiationService) as IInstantiationService;

	const parent = domElement.parentElement;
	const workbench = instantiationService.createInstance(
		Workbench,
		parent,
		domElement,
		{ resource: workspace },
		{},
		services,
	);
	workbench.startup();

	if (Features.zap2Way.isEnabled()) {
		// HACK: overwritten by workbench.startup()
		services.set(ITextModelResolverService, instantiationService.createInstance(TextModelResolverService));
		services.set(ITextFileService, instantiationService.createInstance(GitTextFileService));

		// HACK: get URI's filename in fragment, not in URI path component
		(TextFileEditorModel.prototype as any).getOrCreateMode = function (modeService: IModeService, preferredModeIds: string, firstLineText?: string): TPromise<IMode> {
			return modeService.getOrCreateModeByFilenameOrFirstLine(this.resource.fragment /* file path */, firstLineText); // tslint:disable-line no-invalid-this
		};
	}

	initExtensionHost(workspace);

	instantiationService.invokeFunction(startTappingOutputChannel);

	configurePostStartup(services);
	workbenchListeners.forEach(cb => cb(true));

	return [workbench, services];
}

const workbenchListeners = new Set<(shown: boolean) => void>();

/**
 * onWorkbenchShown registers a listener callback that is invoked whenever
 * a new workbench is bootstrapped.
 */
export function onWorkbenchShown(listener: (shown: boolean) => void): IDisposable {
	workbenchListeners.add(listener);
	return {
		dispose: () => {
			workbenchListeners.delete(listener);
		}
	};
}

/**
 * unmount disposes registered listeners, and should be called when React unmounts
 * the WorkbenchShell component.
 */
export function unmount(): void {
	workbenchListeners.forEach(cb => cb(false));
}
