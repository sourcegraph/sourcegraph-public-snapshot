let load: Promise<void>;

export function loadMonaco(assetsRoot: string): Promise<void> {
	if (!load) {
		load = new Promise((resolve, reject) => {
			let script = document.createElement("script");
			script.type = "text/javascript";
			script.src = `${assetsRoot}/vs/loader.js`;
			script.addEventListener("load", () => {
				(global as any).require.config({ paths: { "vs": `${assetsRoot}/vs` } });
				(global as any).require(["vs/editor/editor.main"], () => resolve());
			});
			document.body.appendChild(script);
		});
	}
	return load;
}
