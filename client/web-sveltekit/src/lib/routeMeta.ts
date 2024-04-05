// A map of route IDs to their corresponding server side route name.
// This feature relies on Vite's glob-import feature
// https://vitejs.dev/guide/features.html#named-imports
// Doing the data processing at runtime is not ideal but by using `import.meta.glob` vite will
// automatically reload this module when any of these files is changed, added or removed.
const rawRouteMeta = import.meta.glob('../routes/**/+@(page|layout).ts', {
    eager: true,
    query: { meta: true },
    import: '_meta',
})

export interface RouteMeta {
    serverRouteName?: string
    isRepoRoute?: boolean
}

export const routeMeta: Record<string, RouteMeta> = {}

// Process raw route meta data. Convert import pats to route IDs and propaget inherited meta data.
const rawEntries = Object.entries(rawRouteMeta) as [string, RouteMeta][]
for (const [path, meta] of rawEntries) {
    if (path.endsWith('+page.ts')) {
        routeMeta[toID(path)] = { ...meta }
    }
}
for (const [path, meta] of rawEntries) {
    if (path.endsWith('+layout.ts')) {
        const layoutPrefix = toID(path)
        for (const [routeID, pageMeta] of Object.entries(routeMeta)) {
            if (routeID.startsWith(layoutPrefix)) {
                routeMeta[routeID] = { ...pageMeta, ...meta }
            }
        }
    }
}

function toID(path: string): string {
    return path.replace(/^\.\.\/routes(\/.*?)\/?\+(page|layout)\.ts$/, '$1')
}
