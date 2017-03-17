// TODO(uforic): Use the features.tsx in the main repo
const ENABLED = "enabled";

class Feature {

	constructor(private name: string) { }

	public isEnabled(): boolean {
		if (!global.window) {
			return false;
		}
		// if not explicitly enabled/disabled, return true if we have beta enabled
		if (localStorage.getItem(this.name) === null) {
			return true;
		}
		return localStorage[this.name] === ENABLED;
	}

	public enable(): void {
		localStorage[this.name] = ENABLED;
	}

	public disable(): void {
		localStorage[this.name] = "disabled";
	}

	public toggle(): void {
		if (this.isEnabled()) {
			this.disable();
		} else {
			this.enable();
		}
	}
}

export const Features = {
	eventLogDebug: new Feature("event-log-debug"),
};

if (global.window) {
	(window as any).features = Features;
}
