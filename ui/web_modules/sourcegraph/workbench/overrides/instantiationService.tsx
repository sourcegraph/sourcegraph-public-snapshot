import { TPromise } from "vs/base/common/winjs.base";
import { InstantiationService } from "vs/platform/instantiation/common/instantiationService";
import { StringEditor } from "vs/workbench/browser/parts/editor/stringEditor";

import { ExplorerViewlet } from "sourcegraph/workbench/overrides/explorerViewlet";

const modules = {
	ExplorerViewlet,
	StringEditor,
};

// Overrides create instance async. This is required so that Webpack can
// statically bundle the modules.
(InstantiationService.prototype as any)._createInstanceAsync = function (desc: any, args: any): any {
	const _this = this;
	const ctor = modules[desc._ctorName];
	if (!ctor) {
		throw "Dynamic modules must be converted to static modules and included in this file.";
	}
	return new TPromise((complete) => {
		// This needs to be async so that the model provider and language
		// providers can register themselves before the editor loads, otherwise
		// the model will fail to resolve.
		setTimeout(() => {
			complete(_this.createInstance(ctor, args));
		});
	});
};
