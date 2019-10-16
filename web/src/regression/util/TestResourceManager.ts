export type ResourceDestructor = () => Promise<void>

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
     * Destroys the resource.
     */
    destroy: () => Promise<void>
}

/**
 * Tracks resources created by tests. Lets the resource creation and removal logic be stored in one
 * place and for easy resource cleanup at the end of tests. Also prints which resources are created
 * and destroyed in case tests are aborted midway through and manual cleanup is required.
 */
export class TestResourceManager {
    private resources: Resource[] = []

    public add(type: Resource['type'], name: string, destroy: () => Promise<void>): void {
        this.resources.push({ type, name, destroy })
    }

    public async destroyAll(): Promise<void> {
        for (const resource of this.resources) {
            await resource.destroy()
            console.log(`Test resource destroyed: ${resource.type} ${JSON.stringify(resource.name)}`)
        }
    }
}
