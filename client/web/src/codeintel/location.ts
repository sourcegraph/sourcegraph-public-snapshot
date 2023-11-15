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
    /** Invariant: `path` matches the 'file' key in all Locations in `locations` */
    public readonly path: string
    /** Invariant: `precise` matches the 'precise' key in all Locations in `locations` */
    private readonly precise: boolean
    /** Invariant: This array is non-empty */
    public readonly locations: Location[]

    /**
     * Pre-condition: `locations` should be non-empty, and every entry
     * should have the same value for 'file'.
     */
    constructor(locations: Location[]) {
        if (locations.length === 0) {
            throw new Error('pre-condition failure')
        }
        this.path = locations[0].file
        const out = LocationsGroupedByFile.preferPrecise(locations)
        this.locations = out.locs
        this.precise = out.precise
    }

    /**
     * Aggregate information about whether provided locations are search-based
     * or precise.
     *
     * If one attempts to mix them, precise locations will be preferred, and
     * search-based locations will be discarded.
     */
    private static preferPrecise(locations: Location[]): { locs: Location[]; precise: boolean } {
        const path: string = locations[0].file
        let precise: boolean = locations[0].precise
        let finalLocations: Location[] = [locations[0]]
        for (const [i, loc] of locations.entries()) {
            if (i === 0) {
                continue
            }
            if (loc.file !== path) {
                throw new Error('pre-condition failure')
            }
            if (precise && !loc.precise) {
                continue
            }
            if (!precise && loc.precise) {
                precise = true
                finalLocations = [loc]
                continue
            }
            if (loc.precise !== precise) {
                throw new Error('already handled precise same-ness check earlier')
            }
            finalLocations.push(loc)
        }
        return { locs: finalLocations, precise }
    }

    public get quality(): LocationGroupQuality {
        return this.precise ? 'PRECISE' : 'SEARCH-BASED'
    }
}

/**
 * Type to store locations grouped by (repo, file) pairs.
 *
 * This type is specialized for use in the reference panel code.
 * So if a given (repo, file) pair contains both search-based Locations
 * and precise Locations, the search-based Locations are discarded.
 */
export class LocationsGroup {
    /**
     * The total number of locations combined across all groups.
     *
     * This may be less than the number of Locations passed to the constructor,
     * in case there are mixed search-based and precise Locations,
     * or if there are duplicates.
     */
    public readonly locationsCount: number
    /** Invariant: Every Location stored in the group has a distinct URL. */
    private readonly groups: LocationsGroupedByRepo[]

    constructor(locations: Location[]) {
        let locationsCount = 0
        const groups: LocationsGroupedByRepo[] = []

        const urlsSeen = new Set<string>()
        const repoMap = new Map<string, Map<string, Location[]>>()
        for (const loc of locations) {
            if (urlsSeen.has(loc.url)) {
                continue
            }
            urlsSeen.add(loc.url)
            let pathToLocMap = repoMap.get(loc.repo)
            if (!pathToLocMap) {
                pathToLocMap = new Map<string, Location[]>()
                repoMap.set(loc.repo, pathToLocMap)
            }
            const fileLocs = pathToLocMap.get(loc.file)
            if (fileLocs) {
                fileLocs.push(loc)
            } else {
                pathToLocMap.set(loc.file, [loc])
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
                locationsCount += g.locations.length
                perFileLocations.push(g)
            }
            groups.push({ repoName, perFileGroups: perFileLocations })
        }

        this.locationsCount = locationsCount
        this.groups = groups
    }

    public get first(): Location | undefined {
        if (this.locationsCount > 0) {
            return this.groups[0].perFileGroups[0].locations[0]
        }
        return undefined
    }

    public get repoCount(): number {
        return this.groups.length
    }

    public map<T>(callback: (arg0: LocationsGroupedByRepo, arg1: number) => T): T[] {
        return this.groups.map(callback)
    }

    public static empty: LocationsGroup = new LocationsGroup([])

    /**
     * Attempt to combine the existing locations with the new set
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
        for (const group of this.groups) {
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
