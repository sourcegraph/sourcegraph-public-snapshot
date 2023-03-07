import { TestSupport } from './test-support'

// The API surface exported to other extensions.
export class ExtensionApi {
	// Hooks for extension test support. This is only set if the
	// environment contains CODY_TESTING=true . This is only for
	// testing and the API will change.
	public testing: TestSupport | undefined = undefined

	constructor() {
		if (process.env.CODY_TESTING === 'true') {
			console.warn('Setting up testing hooks')
			this.testing = new TestSupport()
			TestSupport.instance = this.testing
		}
	}
}
