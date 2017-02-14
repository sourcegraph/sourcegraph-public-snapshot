import { optimizely } from "sourcegraph/tracking/OptimizelyWrapper";
import { experimentEventNames } from "sourcegraph/util/constants/AnalyticsConstants";

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
		if (!this.optimizelyId) {
			return undefined;
		}
		const variationId = optimizely.getVariationId(this.optimizelyId);
		if (!variationId) {
			return undefined;
		}
		return this.getVariationByOptimizelyId(variationId);
	}

	/**
	 * Bucket the user into a specific variation. This method can used from the browser console `features.experimentManager.getExperimentByName("HomepageCopy").setVariation("Previous")`
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
		if (!this.optimizelyId) {
			console.error(`cannot set variation, optimizely id missing from experiment ${this.name}: ${this.optimizelyId}`);
			return;
		}
		if (!variation.optimizelyId) {
			console.error(`cannot set variation, optimizely id missing from variation ${variation.name}: ${variation.optimizelyId}`);
			return;
		}
		optimizely.setVariation(this.optimizelyId, variation.index);
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

// TODO(uforic): We can probably get rid of this.
class ExperimentManagerClass {

	constructor(private experiments: Experiment[]) {
		linkExperimentIds(this.experiments);
	}

	public getExperimentByName(experimentName: string): Experiment | undefined {
		return this.experiments.find(experiment => experiment.name === experimentName);
	}

	public logEvent(eventLabel: string): void {
		if (!experimentEventNames.has(eventLabel)) {
			return;
		}
		optimizely.logEvent(eventLabel);
	}

}

export const experimentManager = new ExperimentManagerClass(liveExperiments);

function linkExperimentIds(experiments: Experiment[]): void {
	const experimentDataList = optimizely.getExperiments();
	optimizely.getActiveExperimentIds().forEach((experimentId) => {
		const optimizelyExp = experimentDataList[experimentId];
		if (!optimizelyExp) {
			console.error(`Experiment id ${experimentId} is active but not present in optimizely.get('data').`);
			return;
		}
		const existingExperiment = experiments.find((el: Experiment) => el.name === optimizelyExp.name);
		if (!existingExperiment) {
			console.error(`An experiment ${optimizelyExp.name} is defined on Optimizely that does not exist in JavaScript.`);
			return;
		}
		existingExperiment.optimizelyId = experimentId;
		optimizelyExp.variations.forEach((optimizelyVariation) => {
			const variation = existingExperiment.getVariationByName(optimizelyVariation.name);
			if (!variation) {
				console.error(`A variation ${optimizelyVariation.name} on experiment ${existingExperiment.name} is defined on Optimizely that does not exist in JavaScript.`);
				return;
			}
			variation.optimizelyId = optimizelyVariation.id;
		}
		);
	});
}
