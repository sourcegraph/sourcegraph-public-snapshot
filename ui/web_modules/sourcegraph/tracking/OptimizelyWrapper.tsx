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

class OptimizelyWrapper {
	private optimizely: OptimizelyJavaScriptApi | null;
	constructor() {
		if (global && global.window && global.window.optimizely && global.window.optimizely.get && global.window.optimizely.push) {
			this.optimizely = global.window.optimizely;
		}
	}

	logEvent(eventLabel: string): void {
		if (!this.optimizely) {
			return;
		}
		this.optimizely.push({
			"type": "event",
			"eventName": eventLabel,
			"tags": {},
		});
	}

	setUserAttributes(attributes: OptimizelyUserAttributes): void {
		if (!this.optimizely) {
			return;
		}
		this.optimizely.push({
			"type": "user",
			"attributes": attributes,
		});
	}

	// Clear all user-specific information on logout (e.g. for public computers), but retain device id (telligent_duid)
	logout(): void {
		this.setUserAttributes({ email: null, user_id: null, is_employee: null });
	}

	getOptimizelyMetadata(): OptimizelyMetadata | {} {
		if (!this.optimizely) {
			return {};
		}
		// Grab all necessary Optimizely data
		const experiments = this.optimizely.get("data").experiments;
		const state = this.optimizely.get("state");

		const variationIdsMap = state.getVariationMap() || {};
		const activeExperimentIds = state.getActiveExperimentIds() || [];
		const experimentInfo: any[] = [];

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

	setVariation(optimizelyId: string, variationIndex?: number): void {
		if (!this.optimizely) {
			console.error("cannot set variation if optimizely has not loaded");
			return;
		}
		this.optimizely.push({
			"type": "bucketVisitor",
			"experimentId": optimizelyId,
			"variationIndex": variationIndex,
		});
	}

	getVariationId(experimentId: string): string | null {
		if (!this.optimizely) {
			return null;
		}
		const currentVariation = this.optimizely.get("state").getVariationMap()[experimentId] as OptimizelyVariation | undefined;
		if (!currentVariation) {
			console.error(`No variation in data map for experiment ${experimentId}.`);
			return null;
		}
		return currentVariation.id;
	}

	getExperiments(): OptimizelyExperiment[] {
		if (!this.optimizely) {
			return [];
		}
		return this.optimizely.get("data").experiments;
	}

	getActiveExperimentIds(): string[] {
		if (!this.optimizely) {
			return [];
		}
		return this.optimizely.get("state").getActiveExperimentIds();
	}

}

export const optimizely = new OptimizelyWrapper();
