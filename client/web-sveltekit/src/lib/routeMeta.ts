// This feature relies on Vite's glob-import feature
// https://vitejs.dev/guide/features.html#named-imports
// Doing the data processing at runtime is not ideal but by using `import.meta.glob` vite will
// automatically reload this module when any of these files is changed, added or removed.
const rawRouteMeta = import.meta.glob('../routes/**/+@(page|layout).ts', {
    eager: true,
    query: { meta: true },
    import: '_meta',
})

/**
 * Extra meta data associated with a route.
 * Pages inherit meta data from their layout.
 * Layouts inerhit meta data from their parents.
 * This doesn't work for pages/layouts that "break out"
 * (https://kit.svelte.dev/docs/advanced-routing#advanced-layouts-breaking-out-of-layouts)
 * but we are not using those atm.
 */
export interface RouteMeta {
    serverRouteName?: string
    isRepoRoute?: boolean
}

/**
 * Map of route IDs to their meta data.
 */
export const routeMeta: Record<string, RouteMeta> = generateRouteMeta(rawRouteMeta)

/**
 * Process raw route meta data. Convert import pats to route IDs and propaget inherited meta data.
 */
function generateRouteMeta(raw: Record<string, unknown>): Record<string, RouteMeta> {
    const rawEntries = Object.entries(raw) as [string, RouteMeta][]
    const layouts: Record<string, RouteMeta> = {}
    const pageMetas: Record<string, RouteMeta> = {}

    for (const [path, meta] of rawEntries) {
        if (path.endsWith('+layout.ts')) {
            layouts[toID(path)] = meta
        }
    }

    for (const [path, pageMeta] of rawEntries) {
        if (path.endsWith('+page.ts')) {
            const pageID = toID(path)
            const meta: RouteMeta = {}

            for (const layout of getLayoutIDsForID(pageID)) {
                if (layouts[layout]) {
                    Object.assign(meta, layouts[layout])
                }
            }

            Object.assign(meta, pageMeta)
            pageMetas[pageID] = meta
        }
    }

    return pageMetas
}

/**
 * DO NOT USE.
 * Exported for testing only.
 */
export const generateRouteMeta_testOnly = generateRouteMeta

function getLayoutIDsForID(routeID: string): string[] {
    const parts = routeID.split('/')
    // route ids always start with '/'
    return ['/'].concat(parts.map((_, i) => parts.slice(0, i + 1).join('/')))
}

function toID(path: string): string {
    return path.replace(/^\.\.\/routes(\/.*?)\/?\+(page|layout)\.ts$/, '$1')
}
