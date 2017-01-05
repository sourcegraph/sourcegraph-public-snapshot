import * as autobind from "autobind-decorator";

// Set of all Sourcegraph events (specifically, eventLabels) that should be sent to Optimizely.
export const OPTIMIZELY_EVENTS = new Set(["SignupCompleted"]);
// List of all live experiment names (must match names defined in the Optimizely admin console)
export enum ExperimentName { HomepageCopy };

class Variation {
	id: string;
	name: string;
	experiment: Experiment;
	// index reflects the order in which a variation was added to the experiment in the Optimizely console
	index: number;

	constructor(id: string, name: string, experiment: Experiment, index: number) {
		this.id = id;
		this.name = name;
		this.experiment = experiment;
		this.index = index;
	}

	// isA() and isB) are helper methods for two-variant experiments; for larger experiments, use isIndex() or isVariation()
	isA(): boolean { return this.index === 0; }
	isB(): boolean { return this.index === 1; }
	isIndex(index: number): boolean { return this.index === index; }
	isVariation(name: string): boolean { return this.name === name; }
}

class Experiment {
	id: string;
	name: string;
	variations: { [id: string]: Variation } = {};

	constructor(id: string, optimizelyExperimentData: any) {
		this.id = id;
		this.name = optimizelyExperimentData.name;
		optimizelyExperimentData.variations.forEach((variation, i) => {
			this.variations[variation.id] = new Variation(variation.id, variation.name, this, i);
		});
	}

	getVariation(name: string): Variation | null {
		for (let id in this.variations) {
			if (this.variations[id] && this.variations[id].isVariation(name)) {
				return this.variations[id];
			}
		}
		return null;
	}
	setVariation(variationName?: string): void {
		ExperimentManager.setVariation(this.name, variationName);
	}
}

@autobind
export class ExperimentManagerClass {
	_liveExperiments: { [name: string]: Experiment } = {};

	constructor() {
		// Build list of all active experiments
		if (global && global.window && global.window.optimizely) {
			const experimentDataList = global.window.optimizely.get("data").experiments;

			global.window.optimizely.get("state").getActiveExperimentIds().forEach((experimentId) => {
				if (experimentDataList[experimentId]) {
					this._liveExperiments[experimentDataList[experimentId].name] = new Experiment(experimentId, experimentDataList[experimentId]);
				}
			});
		}
	}

	logOptimizelyEvent(eventLabel: string): void {
		if (global && global.window && global.window.optimizely) {
			if (OPTIMIZELY_EVENTS.has(eventLabel)) {
				global.window.optimizely.push({
					"type": "event",
					"eventName": eventLabel,
					"tags": { // any event tags/props go here
					},
				});
			}
		}
	}

	setOptimizelyUserAttributes(attrs: any): void {
		if (global && global.window && global.window.optimizely) {
			global.window.optimizely.push({
				"type": "user",
				"attributes": attrs,
			});
		}
	}

	// Bucket the user into a specific variation. This method can used from the browser console `features.ExperimentManager.setVariation(<experimentName>, <variationName>)`
	// Not passing a variationName will bucket the user into the zero-indexed/first-created variation
	setVariation(experimentName: string, variationName?: string): void {
		if (global && global.window && global.window.optimizely) {
			if (ExperimentName[experimentName] === undefined) {
				console.error("Experiment \"%s\" not found", experimentName);
			}
			const experiment = this._getExperiment(ExperimentName[experimentName]);
			if (experiment) {
				let variationIndex = 0;
				if (variationName !== undefined) {
					const variation = experiment.getVariation(variationName);
					if (variation) {
						variationIndex = variation.index;
					} else {
						console.error("Variation \"%s\" not found in experiment \"%s\"", variationName, experiment.name);
					}
				}
				global.window.optimizely.push({
					"type": "bucketVisitor",
					"experimentId": experiment.id,
					"variationIndex": variationIndex,
				});
			}
		}
	}

	getOptimizelyMetadata(): any {
		if (global && global.window && global.window.optimizely) {
			const optimizely = window["optimizely"];
			if (!optimizely) {
				return;
			}

			// Grab all necessary Optimizely data
			const experiments = optimizely.get("data").experiments;
			const state = optimizely.get("state");

			const variationIdsMap = state.getVariationMap() || {};
			const activeExperimentIds = state.getActiveExperimentIds() || [];
			let experimentInfo = [];

			// Loop over all active experiments - track each experiment
			activeExperimentIds.forEach((id) => {
				if (experiments.hasOwnProperty(id)) {
					const experiment = experiments[id];
					experimentInfo.push({
						experimentId: id,
						name: experiment.name,
						variationId: variationIdsMap[id].id,
						variationName: variationIdsMap[id].name,
					});
				}
			});

			return { experimentInfo: experimentInfo };
		}
	}

	// Clear all user-specific information on logout (e.g. for public computers), but retain device id (telligent_duid)
	logoutOptimizely(): void {
		this.setOptimizelyUserAttributes({ email: null, user_id: null, is_employee: null });
	}

	// Todo (DAN): build a better/cleaner structure for providing experiment-level content and changes
	// This simple switch statement-based method works for a small number of very simple experiments (e.g. homepage copy)
	// but won't scale to bigger/page-wide or multi-page-long experiments
	getExperimentContent(experimentName: ExperimentName): any {
		const experiment = this._getExperiment(experimentName);
		const variation = experiment ? this._getCurrentVariation(experimentName) : null;

		this._logToConsole(experimentName, variation ? variation.name : "", experiment);

		switch (experimentName) {
			case ExperimentName.HomepageCopy:
				if (experiment && variation && variation.isB()) {
					return { title: "Welcome to the global graph of code", subTitle: "Read code on the web with the power of an IDE." };
				} else {
					return { title: "Read code on the web with the power of an IDE", subTitle: "Read code smarter and faster. Get more done." };
				}
			default:
				return null;
		}
	}

	_getExperiment(experimentName: ExperimentName): Experiment | null {
		const nameStr = ExperimentName[experimentName];
		if (this._liveExperiments[nameStr]) { return this._liveExperiments[nameStr]; }
		return null;
	}

	_getCurrentVariation(experimentName: ExperimentName): Variation | null {
		const experiment = this._getExperiment(experimentName);
		if (experiment) {
			const currentVariationObject = global.window.optimizely.get("state").getVariationMap()[experiment.id];
			if (currentVariationObject && currentVariationObject.id && experiment.variations[currentVariationObject.id]) {
				return experiment.variations[currentVariationObject.id];
			}
		}
		return null;
	}

	_logToConsole(experimentName: ExperimentName, variationName: string, experiment: any): void {
		if (global && global.window && global.window.features && global.window.features.experimentLogDebug.isEnabled()) {
			const defaultStyle = "color: #aaa";
			const nameStyle = "font-style: italic; color: #C60";
			if (experiment) {
				console.debug("%cEXPERIMENT %c\"%s\"%c (current variation: %c\"%s\"%c)", defaultStyle, nameStyle, ExperimentName[experimentName], defaultStyle, nameStyle, variationName, defaultStyle, experiment); // tslint:disable-line
			} else {
				console.debug("%cEXPERIMENT %c\"%s\"%c not found", defaultStyle, nameStyle, ExperimentName[experimentName], defaultStyle); // tslint:disable-line
			}
		}
	}
}

export const ExperimentManager = new ExperimentManagerClass();
