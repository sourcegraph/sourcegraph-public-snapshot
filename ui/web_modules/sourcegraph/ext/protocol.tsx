/**
 * InitializationOptions are the options sent by main to the extension
 * host when creating the extension host.
 *
 * TODO(sqs): This info will need to be updated on the extension host
 * when it changes on main, but currently this is only sent at the
 * beginning. We can fix this before removing the features.extensions
 * feature flag.
 */
export interface InitializationOptions {
	/**
	 *   The workspace URI
	 */
	workspace: string;

	/**
	 * Feature flags that should be enabled in the extension host.
	 */
	features: string[];
}
