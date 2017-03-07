// This file is mostly copied from vscode's extHost.contribution.ts
// file. It omits services that we don't use, and it also explicitly
// deletes them from the MainContext and ExtHostContext objects so
// that other code doesn't try to use them.

import { IExtensionService } from "vs/platform/extensions/common/extensions";
import { IConstructorSignature0, IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { Registry } from "vs/platform/platform";
import { ExtHostContext, InstanceCollection, MainContext } from "vs/workbench/api/node/extHost.protocol";
import { Extensions as WorkbenchExtensions, IWorkbenchContribution, IWorkbenchContributionsRegistry } from "vs/workbench/common/contributions";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";

import { MainThreadCommands } from "vs/workbench/api/node/mainThreadCommands";
import { MainThreadConfiguration } from "vs/workbench/api/node/mainThreadConfiguration";
import { MainThreadDiagnostics } from "vs/workbench/api/node/mainThreadDiagnostics";
import { MainThreadDocuments } from "vs/workbench/api/node/mainThreadDocuments";
import { MainThreadEditors } from "vs/workbench/api/node/mainThreadEditors";
import { MainThreadErrors } from "vs/workbench/api/node/mainThreadErrors";
import { MainProcessExtensionService } from "vs/workbench/api/node/mainThreadExtensionService";
import { MainThreadLanguageFeatures } from "vs/workbench/api/node/mainThreadLanguageFeatures";
import { MainThreadLanguages } from "vs/workbench/api/node/mainThreadLanguages";
import { MainThreadMessageService } from "vs/workbench/api/node/mainThreadMessageService";
import { MainThreadOutputService } from "vs/workbench/api/node/mainThreadOutputService";
import { MainThreadQuickOpen } from "vs/workbench/api/node/mainThreadQuickOpen";
import { MainThreadStatusBar } from "vs/workbench/api/node/mainThreadStatusBar";
import { MainThreadStorage } from "vs/workbench/api/node/mainThreadStorage";
import { MainThreadTelemetry } from "vs/workbench/api/node/mainThreadTelemetry";
import { MainThreadTreeExplorers } from "vs/workbench/api/node/mainThreadTreeExplorers";
import { MainThreadWorkspace } from "vs/workbench/api/node/mainThreadWorkspace";

export class ExtHostContribution implements IWorkbenchContribution {

	constructor(
		@IThreadService private threadService: IThreadService,
		@IInstantiationService private instantiationService: IInstantiationService,
		@IExtensionService private extensionService: IExtensionService
	) {
		this.initExtensionSystem();
	}

	public getId(): string {
		return "vs.api.extHost";
	}

	private initExtensionSystem(): void {
		const create = (ctor: IConstructorSignature0<any>): any => {
			const service = this.instantiationService.createInstance(ctor);
			if (ctor === MainThreadMessageService && localStorage.getItem("logExtensionHostCommunication") === null) {
				// only open debugger console in development mode
				(service as MainThreadMessageService).disable();
			} else if (ctor === MainThreadOutputService && localStorage.getItem("logExtensionHostCommunication") === null) {
				// only open debugger console in development mode
				(service as MainThreadOutputService).disable();
			}
			return service;
		};

		delete MainContext.MainThreadTerminalService;
		delete MainContext.MainProcessExtensionService;
		delete ExtHostContext.ExtHostTerminalService;

		// Addressable instances
		const col = new InstanceCollection();
		col.define(MainContext.MainThreadCommands).set(create(MainThreadCommands));
		col.define(MainContext.MainThreadConfiguration).set(create(MainThreadConfiguration));
		col.define(MainContext.MainThreadDiagnostics).set(create(MainThreadDiagnostics));
		col.define(MainContext.MainThreadDocuments).set(create(MainThreadDocuments));
		col.define(MainContext.MainThreadEditors).set(create(MainThreadEditors));
		col.define(MainContext.MainThreadErrors).set(create(MainThreadErrors));
		col.define(MainContext.MainThreadExplorers).set(create(MainThreadTreeExplorers));
		col.define(MainContext.MainThreadLanguageFeatures).set(create(MainThreadLanguageFeatures));
		col.define(MainContext.MainThreadLanguages).set(create(MainThreadLanguages));
		col.define(MainContext.MainThreadMessageService).set(create(MainThreadMessageService));
		col.define(MainContext.MainThreadOutputService).set(create(MainThreadOutputService));
		col.define(MainContext.MainThreadQuickOpen).set(create(MainThreadQuickOpen));
		col.define(MainContext.MainThreadStatusBar).set(create(MainThreadStatusBar));
		col.define(MainContext.MainThreadStorage).set(create(MainThreadStorage));
		col.define(MainContext.MainThreadTelemetry).set(create(MainThreadTelemetry));
		col.define(MainContext.MainThreadWorkspace).set(create(MainThreadWorkspace));
		if (this.extensionService instanceof MainProcessExtensionService) {
			col.define(MainContext.MainProcessExtensionService).set(this.extensionService as MainProcessExtensionService);
		}
		col.finish(true, this.threadService);
	}
}

export function registerContribution(): void {
	Registry.as<IWorkbenchContributionsRegistry>(WorkbenchExtensions.Workbench).registerWorkbenchContribution(
		ExtHostContribution
	);
}
