import { context } from "sourcegraph/app/context";

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
	 * The sequence identifier for an extension host, provided by the main
	 * thread. This allows us to prevent collisions on globally namespaced
	 * handles to e.g. hover provider disposables.
	 */
	seqId: number;

	/**
	 * The workspace URI
	 */
	workspace: string;

	/**
	 * Feature flags that should be enabled in the extension host.
	 */
	features: string[];

	/**
	 * revState is the current revision state at the time of initialization.
	 */
	revState?: {
		zapRef?: string,
		commitID?: string,
		branch?: string
	};

	context: typeof context;

	langs: string[] | undefined;
}
