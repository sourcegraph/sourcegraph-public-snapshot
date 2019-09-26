interface Resource {
    type: string
    name: string
    create: () => Promise<void>
    destroy: () => Promise<void>
}

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
