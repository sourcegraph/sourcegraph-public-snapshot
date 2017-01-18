import { Experiment } from "sourcegraph/util/ExperimentManager";

interface OptimizelyVariation {
	id: string;
	name: string;
}

interface OptimizelyExperiment {
	id: string;
	name: string;
	variations: OptimizelyVariation[];
}

interface OptimizelyUserAttributes {
	email?: string | null;
	user_id?: string | null;
	is_employee?: boolean | null;
}

interface OptimizelyJavaScriptApi {
	get: any;
	push: any;
}

interface OptimizelyMetadata {
	experimentInfo: any[];
}

class OptimizelyApiServiceClass {

	constructor(private optimizely: OptimizelyJavaScriptApi) {
		this.optimizely = optimizely;
	}

	public linkExperimentIds(experiments: Experiment[]): void {
		const experimentDataList = this.optimizely.get("data").experiments as OptimizelyExperiment[];
		this.optimizely.get("state").getActiveExperimentIds().forEach((experimentId) => {
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
			optimizelyExp.variations.forEach((optimizelyVariation: OptimizelyVariation) => {
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

	public logEvent(eventLabel: string): void {
		this.optimizely.push({
			"type": "event",
			"eventName": eventLabel,
			"tags": {},
		});
	}

	public setUserAttributes(attributes: OptimizelyUserAttributes): void {
		this.optimizely.push({
			"type": "user",
			"attributes": attributes,
		});
	}

	// Clear all user-specific information on logout (e.g. for public computers), but retain device id (telligent_duid)
	public logout(): void {
		this.setUserAttributes({ email: null, user_id: null, is_employee: null });
	}

	public getOptimizelyMetadata(): OptimizelyMetadata {
		// Grab all necessary Optimizely data
		const experiments = this.optimizely.get("data").experiments;
		const state = this.optimizely.get("state");

		const variationIdsMap = state.getVariationMap() || {};
		const activeExperimentIds = state.getActiveExperimentIds() || [];
		const experimentInfo = [];

		// Loop over all active experiments - track each experiment
		activeExperimentIds.forEach((id) => {
			const experiment = experiments[id];
			if (experiment) {
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

	public setVariation(optimizelyId: string, variationIndex?: number): void {
		this.optimizely.push({
			"type": "bucketVisitor",
			"experimentId": optimizelyId,
			"variationIndex": variationIndex,
		});
	}

	public getVariationId(experimentId: string): string | null {
		const currentVariation = this.optimizely.get("state").getVariationMap()[experimentId] as OptimizelyVariation | undefined;
		if (!currentVariation) {
			console.error(`No variation in data map for experiment ${experimentId}.`);
			return null;
		}
		return currentVariation.id;
	}
}

function getOptimizelyApiService(): OptimizelyApiServiceClass | null {
	const isOptimizelyInstalled = global && global.window && global.window.optimizely && global.window.optimizely.get && global.window.optimizely.push;
	return isOptimizelyInstalled ? new OptimizelyApiServiceClass(global.window.optimizely) : null;
}

export const optimizelyApiService = getOptimizelyApiService();
