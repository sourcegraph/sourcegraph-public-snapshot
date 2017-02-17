// InstantiationService and others (incl. the simpleWorker) expect to
// be able to dynamically require modules. This is incompatible with
// webpack, so in this file we pre-import everything we could possibly
// need, and then expose a require function that satisfies require
// calls statically.
//
// To add a new static import, add it to the staticImports array AND
// add a require() call below.

const staticImports = {
	"vs/workbench/browser/parts/editor/stringEditor": require("vs/workbench/browser/parts/editor/stringEditor"),
	"vs/workbench/browser/parts/editor/textDiffEditor": require("vs/workbench/browser/parts/editor/textDiffEditor"),
	"vs/workbench/parts/files/browser/explorerViewlet": require("vs/workbench/parts/files/browser/explorerViewlet"),
	"vs/workbench/parts/output/browser/outputPanel": require("vs/workbench/parts/output/browser/outputPanel"),
	"vs/workbench/parts/output/common/outputLinkComputer": require("vs/workbench/parts/output/common/outputLinkComputer"),
	"vs/workbench/parts/files/browser/editors/textFileEditor": require("vs/workbench/parts/files/browser/editors/textFileEditor"),
	"vs/workbench/parts/files/common/editors/fileEditorInput": require("vs/workbench/parts/files/common/editors/fileEditorInput"),
};

window["require"] = (modules: string[], callback: (mod: any) => void): void => {
	if (modules.length !== 1) { throw new Error(`expected 1 module: ${JSON.stringify(modules)}`); }
	const moduleName = modules[0];
	if (!staticImports[moduleName]) {
		throw new Error(`Module ${moduleName} was dynamically required, but webpack does not support dynamic requires. It must be added to the staticImport objects in staticImports.tsx.`);
	}
	require.ensure([], require => { // tslint:disable-line no-shadowed-variable
		callback(staticImports[moduleName]);
	});
};
