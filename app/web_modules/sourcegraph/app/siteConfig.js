// @flow weak

import * as React from "react";
import EventLogger from "sourcegraph/util/EventLogger";

// SiteConfig is site-wide configuration for a Sourcegraph server.
export type SiteConfig = {
	appURL: string; // base URL for app (e.g., https://sourcegraph.com or http://localhost:3080)
	assetsRoot: string; // URL path to image/font/etc. assets on server
	buildVars: Object; // from the build process (sgtool)
};

let _globalSiteConfig: ?SiteConfig = null; // private, access via withSiteConfigContext and this.context.siteConfig

// setGlobalSiteConfig sets the feature flags that will be provided to
// React components via the withSiteConfigContext wrapper and
// "this.context.siteConfig" context item.
//
// This module assumes that the siteConfig object is immutable
// and it and its subkeys will not change. Violating this will result in
// undefined behavior.
export function setGlobalSiteConfig(siteConfig: SiteConfig) {
	if (_globalSiteConfig !== null) throw new Error("global feature flags already set, may only be set once");
	_globalSiteConfig = siteConfig;

	// HACK: Set this info in other places that need it but that
	// aren't React components (which could access it via this.context.siteConfig).
	EventLogger.setSiteConfig(siteConfig);
}

// withStatusContext passes a "siteConfig" context item
// to Component's children and lets them query the
// feature flags.
export function withSiteConfigContext(Component) {
	class WithSiteConfig extends React.Component {
		static childContextTypes = {
			siteConfig: React.PropTypes.object,
		};

		getChildContext(): {siteConfig: SiteConfig} {
			if (!_globalSiteConfig) throw new Error("siteConfig not set");
			return {siteConfig: _globalSiteConfig};
		}

		render() {
			return <Component {...this.props} />;
		}
	}
	return WithSiteConfig;
}
