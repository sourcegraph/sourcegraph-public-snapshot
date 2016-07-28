// @flow weak

import * as React from "react";

// FeatureName is the set of all feature flag names that can
// be queried by the frontend app. This list is usually a subset
// of the feature flags in the ./conf/feature package.
export type FeatureName =
	"Authors" |
	"GodocRefs" |
	"_testingDummyFeature"; // used by tests only

export type Features = {[key: FeatureName]: any};

let _globalFeatures: ?Features = null; // private, access via withFeaturesContext and this.context.features

// setGlobalFeatures sets the feature flags that will be provided to
// React components via the withFeaturesContext wrapper and
// "this.context.features" context item.
//
// This module assumes that the features object is immutable
// and it and its subkeys will not change. Violating this will result in
// undefined behavior.
export function setGlobalFeatures(features: Features) {
	_globalFeatures = features;
}

// withFeaturesContext passes a "features" context item
// to Component's children and lets them query the
// feature flags.
export function withFeaturesContext(Component) {
	class WithFeatures extends React.Component {
		static childContextTypes = {
			features: React.PropTypes.object,
		};

		getChildContext(): {features: Features} {
			if (!_globalFeatures) throw new Error("features not set");
			return {features: _globalFeatures};
		}

		render() {
			return <Component {...this.props} />;
		}
	}
	return WithFeatures;
}
