import { Text } from '@sourcegraph/wildcard'

const COLORS = [
    'var(--oc-violet-0)',
    'var(--oc-violet-1)',
    'var(--oc-violet-2)',
    'var(--oc-violet-3)',
    'var(--oc-violet-4)',
    'var(--oc-violet-5)',
    'var(--oc-violet-6)',
    'var(--oc-violet-7)',
    'var(--oc-violet-8)',
    'var(--oc-violet-9)',
]
const DARK_COLORS = COLORS.slice(0).reverse()

const ONE_YEAR_AGO = Date.now() - 1000 * 60 * 60 * 24 * 365

export function useBlameRecencyColor(
    commit?: Date,
    // @TODO: Pass actual repo creation date
    creation?: Date
): string {
    // @TODO: Pass through the actual flag
    const isLightTheme = false
    const colors = isLightTheme ? COLORS : DARK_COLORS

    if (!commit) {
        return colors[0]
    }
    if (!creation) {
        creation = new Date(Date.now() - 3 * 1000 * 60 * 60 * 24 * 365)
    }

    // We create a recency range depending on the repo creation date. If the
    // repo is newer than a year, we use the last year so that we don't have a
    // scale that is too sensible.
    const now = Date.now()
    const start = Math.min(creation.getTime(), ONE_YEAR_AGO)

    // We should probably not use a linear scale here :shrug:
    const recency = Math.min(Math.max((now - commit.getTime()) / (now - start), 0), 1)

    return colors[Math.ceil(recency * 10) - 1]
}
