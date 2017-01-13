import * as optimizely from "sourcegraph/util/Optimizely";

export class Variation {
	public optimizelyId?: string;

	constructor(public index: number, public name: string) { }

	// isA() and isB) are helper methods for two-variant experiments; for larger experiments, use isIndex() or isVariation()
	public isA(): boolean { return this.index === 0; }
	public isB(): boolean { return this.index === 1; }
	public isIndex(index: number): boolean { return this.index === index; }
	public isVariation(name: string): boolean { return this.name === name; }
}

export class Experiment {
	public optimizelyId?: string;
	public getContent: () => Object;

	constructor(public name: string, public variations: Variation[], getContent: (variation: Variation | undefined) => Object) {
		this.getContent = () => {
			this.logToConsole();
			return getContent(this.getCurrentVariation());
		};
	}

	public getVariationByName(name: string): Variation | undefined {
		return this.variations.find((variation: Variation) => variation.name === name);
	}

	public getVariationByOptimizelyId(optimizelyId: string): Variation | undefined {
		return this.variations.find((variation) => variation.optimizelyId === optimizelyId);
	}

	private getCurrentVariation(): Variation | undefined {
		if (!(optimizely.optimizelyApiService && this.optimizelyId)) {
			return undefined;
		}
		const variationId = optimizely.optimizelyApiService.getVariationId(this.optimizelyId);
		if (!variationId) {
			return undefined;
		}
		return this.getVariationByOptimizelyId(variationId);
	}

	/**
	 * Bucket the user into a specific variation. This method can used from the browser console `features.ExperimentManager.HomepageCopy.setVariation(<variationName>)`
	 */
	public setVariation(variationName?: string): void {
		let variation: Variation | undefined;
		if (variationName) {
			variation = this.getVariationByName(variationName);
		}
		if (!variation) {
			console.error(`Variation ${variationName} not found in experiment ${this.name}.`);
			return;
		}
		if (!optimizely.optimizelyApiService) {
			return;
		}
		if (!this.optimizelyId) {
			console.error(`cannot set variation, optimizely id missing from experiment ${this.name}: ${this.optimizelyId}`);
			return;
		}
		if (!variation.optimizelyId) {
			console.error(`cannot set variation, optimizely id missing from variation ${variation.name}: ${variation.optimizelyId}`);
			return;
		}
		optimizely.optimizelyApiService.setVariation(this.optimizelyId, variation.optimizelyId);
	}

	private logToConsole(): void {
		const variation = this.getCurrentVariation();
		if (!(global && global.window && global.window.features && global.window.features.experimentLogDebug.isEnabled())) {
			return;
		}
		const defaultStyle = "color: #aaa";
		const nameStyle = "font-style: italic; color: #C60";
		console.debug("%cEXPERIMENT %c\"%s\"%c (current variation: %c\"%s\"%c)", defaultStyle, nameStyle, this.name, defaultStyle, nameStyle, variation ? variation.name : "undefined", defaultStyle, this); // tslint:disable-line
	}
}

export interface HomepageExperimentContent {
	title: string;
	subTitle: string;
}

export const homepageExperiment = new Experiment("HomepageCopy", [new Variation(0, "Current"), new Variation(1, "Previous")], (variation) => {
	if (variation && variation.isB()) {
		return { title: "Welcome to the global graph of code", subTitle: "Read code on the web with the power of an IDE." };
	}
	return { title: "Read code on the web with the power of an IDE", subTitle: "Read code smarter and faster. Get more done." };
});

export const liveExperiments = [homepageExperiment];

// Set of all Sourcegraph events (specifically, eventLabels) that should be sent to Optimizely.
const experimentEventNames = new Set("SignupCompleted");

// TODO(uforic): We can probably get rid of this.
class ExperimentManagerClass {

	constructor(private experiments: Experiment[]) {
		if (optimizely.optimizelyApiService) {
			optimizely.optimizelyApiService.linkExperimentIds(this.experiments);
		}
	}

	public logEvent(eventLabel: string): void {
		if (!experimentEventNames.has(eventLabel)) {
			return;
		}
		if (!optimizely.optimizelyApiService) {
			return;
		}
		optimizely.optimizelyApiService.logEvent(eventLabel);
	}

}

export const experimentManager = new ExperimentManagerClass(liveExperiments);
