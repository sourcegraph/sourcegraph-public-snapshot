import { Permutations } from './types'
import * as packages from './packages.json'

// @ts-ignore
let PACKAGE_NAMES = new Set(packages.default.slice(0, 100).map(pkg => pkg.name))

const alphanumericSort = (a: string, b: string) => a.localeCompare(b, 'en', { numeric: true })

export function createPermutations(): Permutations {
    let permutations: Permutations = new Map()

    for (const _a of PACKAGE_NAMES) {
        for (const _b of PACKAGE_NAMES) {
            if (_a === _b) {
                continue
            }

            const [a, b]: [string, string] = [_a as any, _b as any].sort(alphanumericSort) as any

            if (!permutations.has(a)) {
                permutations.set(a, new Set())
            }

            permutations.get(a)?.add(b)
        }
    }

    return permutations
}
