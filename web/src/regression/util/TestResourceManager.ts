export type ResourceDestructor = () => Promise<void>

interface Resource {
    /**
     * Resource type, printed on creation and destruction.
     */
    type: 'User' | 'External service' | 'Authentication provider' | 'Global setting' | 'Organization' | 'Configuration'

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

    public add<T>(type: Resource['type'], name: string, value: { result: T; destroy: () => Promise<void> }): T
    public add(type: Resource['type'], name: string, destroy: () => Promise<void>): void
    public add(type: Resource['type'], name: string, v: any): any {
        if (v.destroy) {
            this.resources.push({ type, name, destroy: v.destroy })
            return v.result
        }
        this.resources.push({ type, name, destroy: v })
    }

    public async destroyAll(): Promise<void> {
        for (const resource of this.resources) {
            await resource.destroy()
            console.log(`Test resource destroyed: ${resource.type} ${JSON.stringify(resource.name)}`)
        }
    }
}
