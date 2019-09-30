interface Resource {
    type: string
    name: string
    create: () => Promise<void>
    destroy: () => Promise<void>
}

/**
 * Tracks resources created by tests. Lets the resource creation and removal logic be stored in one
 * place and for easy resource cleanup at the end of tests. Also prints which resources are created
 * and destroyed in case tests are aborted midway through and manual cleanup is required.
 */
export class TestResourceManager {
    private resources: Resource[]

    constructor(resources?: Resource[]) {
        this.resources = resources || []
    }

    public async create(resource: Resource): Promise<void> {
        await resource.create()
        this.resources.push(resource)
        console.log(`Test resource created: ${resource.type} ${JSON.stringify(resource.name)}`)
    }

    public async destroyAll(): Promise<void> {
        for (const resource of this.resources) {
            await resource.destroy()
            console.log(`Test resource destroyed: ${resource.type} ${JSON.stringify(resource.name)}`)
        }
    }
}
