import { TPromise } from "vs/base/common/winjs.base";
import { ExtensionPointContribution, IExtensionDescription, IExtensionService, IExtensionsStatus } from "vs/platform/extensions/common/extensions";
import { IExtensionPoint } from "vs/platform/extensions/common/extensionsRegistry";

/**
 * ExtensionService is a minimal implementation of IExtensionService.
 */
export class ExtensionService implements IExtensionService {
	_serviceBrand: any;
	activateByEvent(activationEvent: string): TPromise<void> { return TPromise.wrapError<void>(new Error("not implemented")); }
	onReady(): TPromise<boolean> { return TPromise.as(true); }
	getExtensions(): TPromise<IExtensionDescription[]> { return TPromise.as([]); }
	readExtensionPointContributions<T>(extPoint: IExtensionPoint<T>): TPromise<ExtensionPointContribution<T>[]> { return TPromise.as<ExtensionPointContribution<T>[]>([]); }
	getExtensionsStatus(): { [id: string]: IExtensionsStatus } { return {}; }
}
