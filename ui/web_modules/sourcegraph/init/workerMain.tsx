// Adapted from vs/base/worker/workerMain.

// This is a mapping of ALL dynamic imports that our Web Worker will
// ever need to perform. This is necessary because for some reason in
// the source of imported modules, `require.ensure` does not get
// rewritten by webpack for some reason.
//
// This is a Sourcegraph customization and must be used explicitly;
// anywhere you see self.require, it is pulling from this map.
// Anywhere else you see require or import, it's standard webpack.
const staticRequires = {
	"vs/editor/common/services/editorSimpleWorker": require("vs/editor/common/services/editorSimpleWorker"),
	"vs/base/common/worker/simpleWorker": require("vs/base/common/worker/simpleWorker"),
	"vs/workbench/parts/output/common/outputLinkComputer": require("vs/workbench/parts/output/common/outputLinkComputer"),
};
// Make self.require callable without clobbering self.require.config.
self["require"] = function (moduleIds: string[], callback: (mod: any) => void): void {
	if (moduleIds.length !== 1) {
		throw new Error("not yet implemented: require of multiple modules: " + moduleIds.join(", "));
	}
	const m = staticRequires[moduleIds[0]];
	if (!m) {
		throw new Error(`unable to load module ${moduleIds[0]} - you must add it to the staticRequires map in workerMain.tsx`);
	}
	callback(m);
};

self["require"].config = () => { /* noop */ };

function loadCode(moduleId: string): void {
	self["require"](["vs/base/common/worker/simpleWorker"], ws => {
		setTimeout(() => {
			const messageHandler = ws.create(msg => {
				((self as any) as Worker).postMessage(msg, []);
			}, null);
			self.onmessage = e => messageHandler.onmessage(e.data);
			while (beforeReadyMessages.length > 0) {
				self.onmessage(beforeReadyMessages.shift());
			}
		}, 0);
	});
};

let isFirstMessage = true;
const beforeReadyMessages: any[] = [];
self.onmessage = message => {
	if (!isFirstMessage) {
		beforeReadyMessages.push(message);
		return;
	}
	isFirstMessage = false;
	loadCode(message.data);
};
