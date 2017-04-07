import { experimentManager } from "sourcegraph/util/ExperimentManager";
import isWebWorker from "sourcegraph/util/isWebWorker";

const enabled = "enabled";

/**
 * storage is localStorage in the main thread and is
 * temporary/in-memory in Web Workers.
 */
let storage: {
	getItem(name: string): string | null;
	setItem(name: string, value: string): void;
};
if (isWebWorker) {
	const data = new Map<string, string>();
	storage = {
		getItem(name: string): string | null {
			const value = data.get(name);
			if (value === undefined) { return null; }
			return value;
		},
		setItem(name: string, value: string): void {
			data.set(name, value);
		},
	};
} else {
	storage = localStorage;
}

class Feature {
	private beta: boolean = true;

	constructor(private name: string) { }

	public isEnabled(): boolean {
		if (!storage) { return false; }

		// if not explicitly enabled/disabled, return true if we have beta enabled
		if (this.beta && storage.getItem(this.name) === null && Features.beta.isEnabled()) {
			return true;
		}
		return storage.getItem(this.name) === enabled;
	}

	public enable(): void {
		storage.setItem(this.name, enabled);
	}

	public disable(): void {
		storage.setItem(this.name, "disabled");
	}

	public toggle(): void {
		if (this.isEnabled()) {
			this.disable();
		} else {
			this.enable();
		}
	}

	public disableBeta(): this {
		this.beta = false;
		return this;
	}
}

export const Features = {
	langCSS: new Feature("lang-css").disableBeta(),
	langPHP: new Feature("lang-php"),
	langPython: new Feature("lang-python").disableBeta(),

	/**
	 * Enable logging for references
	 */
	refLogs: new Feature("refLogs"),

	/**
	 * trace is whether to show trace URLs to LightStep in console log messages.
	 */
	trace: new Feature("trace"),

	zap: new Feature("zap").disableBeta(),
	zap2Way: new Feature("zap-2-way").disableBeta(),

	beta: new Feature("beta").disableBeta(),
	eventLogDebug: new Feature("event-log-debug").disableBeta(),
	actionLogDebug: new Feature("action-log-debug").disableBeta(),
	experimentLogDebug: new Feature("experiment-log-debug").disableBeta(),
	zapSelections: new Feature("zapSelections").disableBeta(),
	zapChanges: new Feature("zapChanges").disableBeta(),

	experimentManager,
	listEnabled,
};

if (Features.zap2Way.isEnabled()) {
	// zap2Way requires different CSS.
	if (document) {
		document.body.classList.add("zap2Way");
	}
}

export function listEnabled(): string[] {
	return Object.keys(Features).filter(name => Features[name] instanceof Feature && Features[name].isEnabled());
}

export function bulkEnable(featureNames: string[]): void {
	for (const name of featureNames) {
		Features[name].enable();
	}
}

export function getModes(onPrem?: boolean): Set<string> {
	let modes = new Set<string>(["go", "java", "javascript", "typescript"]);
	if (onPrem) {
		modes = new Set<string>(["go", "java"]);
	}
	if (Features.langCSS.isEnabled()) {
		modes.add("css");
		modes.add("less");
		modes.add("scss");
	}
	if (Features.langPHP.isEnabled()) {
		modes.add("php");
	}
	if (Features.langPython.isEnabled()) {
		modes.add("python");
	}
	return modes;
}

if (global.window) {
	(window as any).features = Features;

	// Make it so that visiting https://sourcegraph.com/#feature=NAME
	// automatically enables the NAME feature in the user's
	// localStorage, so we can share beta links with external users
	// more easily.
	if (document.location.hash) {
		const m = document.location.hash.match(/^#feature=(.*)/);
		if (m) {
			const name = m[1];
			Features[name].enable();
			console.log("Enabling feature flag:", name); // tslint:disable-line no-console
			document.location.hash = "";
		}
	}
}
