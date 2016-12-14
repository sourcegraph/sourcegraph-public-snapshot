import { TPromise } from "vs/base/common/winjs.base";
import { InstantiationService } from "vs/platform/instantiation/common/instantiationService";
import { StringEditor } from "vs/workbench/browser/parts/editor/stringEditor";
import { ExplorerViewlet } from "vs/workbench/parts/files/browser/explorerViewlet";

const modules = {
	ExplorerViewlet,
	StringEditor,
};

(InstantiationService.prototype as any)._createInstanceAsync = function(desc: any, args: any) {
	const _this = this;
	const ctor = modules[desc._ctorName];
	return new TPromise((c, e) => {
		setTimeout(() => {
			c(_this.createInstance(ctor, args))
		});
	});
};
