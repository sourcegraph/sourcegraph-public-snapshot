import type { Range } from '@sourcegraph/extension-api-types'

import type { LocationFields } from '../graphql-operations'

import type { Result } from './searchBased'

export interface Location {
    repo: string
    file: string
    content: string
    commitID: string
    range?: Range
    url: string
    lines: string[]
    precise: boolean
}

export interface LocationsGroupedByRepo {
    /** Invariant: `repoName` matches the 'repo' key in all Locations in `perFileGroups` */
    repoName: string
    /** Invariant: This array is non-empty */
    perFileGroups: LocationsGroupedByFile[]
}

export type LocationGroupQuality = 'PRECISE' | 'SEARCH-BASED'

export class LocationsGroupedByFile {
    /** Invariant: `path` matches the 'file' key in all Locations in `_locations` */
    public path: string
    /** Invariant: `precise` matches the 'precise' key in all Locations in `_locations` */
    private _precise: boolean
    /** Invariant: This array is non-empty */
    private _locations: Location[]

    /** Pre-condition: `locations` should be non-empty, and every entry
     * should have the same value for 'file'.
     */
    constructor(locations: Location[]) {
        if (locations.length === 0) {
            throw new Error('pre-condition failure')
        }
        this.path = locations[0].file
        this._precise = locations[0].precise
        this._locations = [locations[0]]
        for (const [i, loc] of locations.entries()) {
            if (i === 0) {
                continue
            }
            this.tryAdd(loc)
        }
    }

    /** Attempt to add the Location to this group without mixing
     * search-based and precise Locations.
     *
     * If one attempts to mix them, precise locations will be preferred.
     */
    private tryAdd(location: Location): void {
        if (this._precise && !location.precise) {
            return
        }
        if (!this._precise && location.precise) {
            this._precise = true
            this._locations = [location]
            return
        }
        if (location.file !== this.path) {
            throw new Error('pre-condition failure')
        }
        if (location.precise !== this._precise) {
            throw new Error('already handled precise same-ness check earlier')
        }
        this._locations.push(location)
    }

    public get locations(): Location[] {
        return this._locations
    }

    public get quality(): LocationGroupQuality {
        return this._precise ? 'PRECISE' : 'SEARCH-BASED'
    }
}

/** Type to store locations grouped by (repo, file) pairs.
 *
 * This type is specialized for use in the reference panel code.
 * So if a given (repo, file) pair contains both search-based Locations
 * and precise Locations, the search-based Locations are discarded.
 * */
export class LocationsGroup {
    /** Invariant: `_totalCount` is the sum of sizes of Location arrays in `_groups` */
    private _locationsCount: number
    /** Invariant: Every Location stored in the group has a distinct URL. */
    private _groups: LocationsGroupedByRepo[]

    constructor(locations: Location[]) {
        this._locationsCount = 0
        this._groups = []

        const urlsSeen = new Set<string>()
        const repoMap = new Map<string, Map<string, Location[]>>()
        for (const loc of locations) {
            if (urlsSeen.has(loc.url)) {
                continue
            }
            urlsSeen.add(loc.url)
            const pathToLocMap = repoMap.get(loc.repo)
            if (pathToLocMap) {
                const fileLocs = pathToLocMap.get(loc.file)
                if (fileLocs) {
                    fileLocs.push(loc)
                } else {
                    pathToLocMap.set(loc.file, [loc])
                }
            } else {
                const pathToLocMap = new Map<string, Location[]>()
                pathToLocMap.set(loc.file, [loc])
                repoMap.set(loc.repo, pathToLocMap)
            }
        }
        for (const [repoName, pathToLocMap] of repoMap) {
            const perFileLocations: LocationsGroupedByFile[] = []
            for (const [path, locations] of pathToLocMap) {
                if (locations.length === 0) {
                    throw new Error(
                        `bug in grouping logic created empty locations array for repo: ${repoName}, path: ${path}`
                    )
                }
                const g = new LocationsGroupedByFile(locations)
                if (g.locations.length > locations.length) {
                    throw new Error('materialized new locations out of thin air')
                }
                this._locationsCount += g.locations.length
                perFileLocations.push(g)
            }
            this._groups.push({ repoName, perFileGroups: perFileLocations })
        }
    }

    /** Returns the total number of locations combined across all groups.
     *
     * This may be less than the number of Locations passed to the constructor,
     * in case there are mixed search-based and precise Locations,
     * or if there are duplicates.
     */
    public get locationsCount(): number {
        return this._locationsCount
    }

    public get first(): Location | undefined {
        if (this._locationsCount > 0) {
            return this._groups[0].perFileGroups[0].locations[0]
        }
        return undefined
    }

    public get repoCount(): number {
        return this._groups.length
    }

    public map<T>(callback: (arg0: LocationsGroupedByRepo, arg1: number) => T): T[] {
        return this._groups.map(callback)
    }

    public static empty: LocationsGroup = new LocationsGroup([])

    /** Attempt to combine the existing locations with the new set
     * into a new LocationsGroup.
     *
     * Some of the Locations in `newLocations` may be dropped if they
     * are search-based and we already had precise references for the
     * same file.
     */
    public combine(newLocations: Location[]): LocationsGroup {
        return new LocationsGroup([...this.allLocations(), ...newLocations])
    }

    private allLocations(): Location[] {
        const out: Location[] = []
        for (const group of this._groups) {
            for (const locs of group.perFileGroups) {
                out.push(...locs.locations)
            }
        }
        return out
    }
}

export const buildSearchBasedLocation = (node: Result): Location => ({
    repo: node.repo,
    file: node.file,
    commitID: node.rev,
    content: node.content,
    url: node.url,
    lines: split(node.content),
    precise: false,
    range: node.range,
})

export const split = (content: string): string[] => content.split(/\r?\n/)

export const buildPreciseLocation = (node: LocationFields): Location => {
    const location: Location = {
        content: node.resource.content,
        commitID: node.resource.commit.oid,
        repo: node.resource.repository.name,
        file: node.resource.path,
        url: node.url,
        lines: [],
        precise: true,
    }
    if (node.range !== null) {
        location.range = node.range
    }
    location.lines = location.content.split(/\r?\n/)
    return location
}
