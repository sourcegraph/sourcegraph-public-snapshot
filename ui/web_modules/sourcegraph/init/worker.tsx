window["MonacoEnvironment"] = {
	getWorkerUrl(workerId: string, label: string): string {
		// Prepare the JavaScript source code that we eval in order to run
		// workerMain in a Web Worker.
		let source = (require as any)("raw-loader!inline-worker-loader.js?inline!sourcegraph/init/workerMain");

		// Import scripts from webpack assets directory.
		source = source.replace("importScripts(\"\"", `importScripts(${JSON.stringify(document.head.dataset["webpackPublicPath"])}`);

		// Provide dummy require options for the vscode loader.
		if (!window["require"]) {
			window["require"] = {};
		}
		Object.assign(window["require"], {
			config(config: any): void { /* noop */ },
			getConfig(): any { return {}; },
		});

		// Return a blob URL containing the JavaScript source code
		// that runs this Web Worker (plus add a dummy
		// require.config for the vscode loader).
		return makeBlobURL(source);
	}
};

// http://stackoverflow.com/questions/10343913/how-to-create-a-web-worker-from-a-string
const windowURL: any = (window as any).URL || (window as any).webkitURL;
export function makeBlobURL(content: string): string {
	return windowURL.createObjectURL(new Blob([content]));
}
