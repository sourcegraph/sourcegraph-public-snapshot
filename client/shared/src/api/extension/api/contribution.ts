import { mapValues } from 'lodash'

import { ContributableMenu, Contributions, MenuItemContribution } from '@sourcegraph/client-api'

/**
 * Merges the contributions.
 *
 * Most callers should use ContributionRegistry#getContributions, which merges all registered
 * contributions.
 */
export function mergeContributions(contributions: Contributions[]): Contributions {
    if (contributions.length === 0) {
        return {}
    }
    if (contributions.length === 1) {
        return contributions[0]
    }
    const merged: Contributions = {}
    for (const contribution of contributions) {
        // swallow errors from malformed manifests to prevent breaking other
        // contributions or extensions: https://github.com/sourcegraph/sourcegraph/pull/12573
        if (contribution.actions) {
            try {
                if (!merged.actions) {
                    merged.actions = [...contribution.actions]
                } else {
                    merged.actions = [...merged.actions, ...contribution.actions]
                }
            } catch {
                // noop
            }
        }
        if (contribution.menus) {
            try {
                if (!merged.menus) {
                    merged.menus = { ...contribution.menus }
                } else {
                    for (const [menu, items] of Object.entries(contribution.menus) as [
                        ContributableMenu,
                        MenuItemContribution[]
                    ][]) {
                        const mergedItems = merged.menus[menu]
                        try {
                            if (!mergedItems) {
                                merged.menus[menu] = [...items]
                            } else {
                                merged.menus[menu] = [...mergedItems, ...items]
                            }
                        } catch {
                            // noop
                        }
                    }
                }
            } catch {
                // noop
            }
        }
        if (contribution.views) {
            try {
                if (!merged.views) {
                    merged.views = [...contribution.views]
                } else {
                    merged.views = [...merged.views, ...contribution.views]
                }
            } catch {
                // noop
            }
        }
        if (contribution.searchFilters) {
            try {
                if (!merged.searchFilters) {
                    merged.searchFilters = [...contribution.searchFilters]
                } else {
                    merged.searchFilters = [...merged.searchFilters, ...contribution.searchFilters]
                }
            } catch {
                // noop
            }
        }
    }
    return merged
}

/**
 * Filters the contributions to only those that are enabled in the current context.
 */
export function filterContributions(contributions: Contributions): Contributions {
    if (!contributions.menus) {
        return contributions
    }
    return {
        ...contributions,
        menus: mapValues(contributions.menus, menuItems => menuItems?.filter(menuItem => menuItem.when !== false)),
    }
}
