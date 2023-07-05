export interface TreeProvider<T> {
    getEntries(): T[]
    getDisplayName(entry: T): string
    getSVGIconPath(entry: T, open: boolean): string
    canOpen(entry: T): boolean
    fetchChildren(entry: T): Promise<TreeProvider<T>>
    getURL(entry: T): string | null
    getKey(entry: T): string
    markOpen(entry: T, open: boolean): void
    isOpen(entry: T): boolean
}

export class DummyTreeProvider implements TreeProvider<any> {
    getDisplayName(entry: any): string {
        return ''
    }
    canOpen(entry: any): boolean {
        return false
    }
    getSVGIconPath(entry: any, open: boolean): string {
        return ''
    }
    getEntries(): any[] {
        return []
    }
    fetchChildren(entry: any): Promise<TreeProvider<any>> {
        return Promise.resolve(new this.constructor())
    }
    getURL(entry: any): string | null {
        return null
    }
    getKey(entry: any): string {
        return ''
    }
}
