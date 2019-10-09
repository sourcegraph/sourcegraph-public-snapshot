interface Resource {
    /**
     * Resource type, printed on creation and destruction.
     */
    type: 'User' | 'External service' | 'Authentication provider'

    /**
     * Name of the resource, printed upon creation and destruction.
     */
    name: string

    /**
     * Creates the resource and returns the destructor for the resource.
     */
    create: () => Promise<() => Promise<void>>
}

/**
 * Tracks resources created by tests. Lets the resource creation and removal logic be stored in one
 * place and for easy resource cleanup at the end of tests. Also prints which resources are created
 * and destroyed in case tests are aborted midway through and manual cleanup is required.
 */
export class TestResourceManager {
    private resources: [Resource, () => Promise<void>][] = []

    public async create(resource: Resource): Promise<void> {
        const destroy = await resource.create()
        this.resources.push([resource, destroy])
        console.log(`Test resource created: ${resource.type} ${JSON.stringify(resource.name)}`)
    }

    public async destroyAll(): Promise<void> {
        for (const [resource, destroy] of this.resources) {
            await destroy()
            console.log(`Test resource destroyed: ${resource.type} ${JSON.stringify(resource.name)}`)
        }
    }
}
