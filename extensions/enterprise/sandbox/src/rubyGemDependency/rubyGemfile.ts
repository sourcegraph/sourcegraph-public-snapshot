import { BundlerExecutionContext, executeBundlerCommand } from './rubyBundler'
import { sortBy } from 'lodash'
import * as sourcegraph from 'sourcegraph'

export interface RubyGemfileDependency {
    name: string
    version?: string

    direct: boolean

    /**
     * The character range in the Gemfile (TODO!(sqs) or Gemfile.lock, distinguish between these 2) where this dependency is defined (for direct dependencies only).
     */
    range?: sourcegraph.Range

    /**
     * The direct dependencies (gem names) that cause this dependency to be included (may be empty
     * for top-level dependencies).
     */
    directAncestors: string[]
}

/** Return canonical sort order and format. */
const toDependencyList = (m: Map<string, RubyGemfileDependency>): RubyGemfileDependency[] => {
    const sorted = sortBy(Array.from(m.values()), d => d.name)
    for (const d of sorted) {
        d.directAncestors.sort()
    }
    return sorted
}

export const parseRubyGemfileLock = (gemfileLock: string): RubyGemfileDependency[] => {
    const byName = new Map<string, RubyGemfileDependency>()

    /** The last-seen parent dependency (whose dependencies we're enumerating). */
    let inGemSection = false
    let inGemSpecsSection = false
    let parentDependency: RubyGemfileDependency | undefined = undefined
    for (const [i, line] of gemfileLock.split(/\n+/).entries()) {
        if (line === '' || line.startsWith('#')) {
            continue
        }
        if (line.startsWith('GEM')) {
            inGemSection = true
            continue
        }
        if (!line.startsWith(' ')) {
            if (inGemSection) {
                break // finished with GEM section
            }
            continue
        }
        if (!inGemSection) {
            continue
        }

        if (line.startsWith('  specs:')) {
            inGemSpecsSection = true
            continue
        }
        if (!inGemSpecsSection) {
            continue
        }

        const DEP_INDENT = '    '
        const TRANSITIVE_DEP_INDENT = '      '
        if (line.startsWith(DEP_INDENT) && !line.startsWith(TRANSITIVE_DEP_INDENT)) {
            // Dependency.
            const m = line.match(/^\s*(.+) \((.+)\)$/)
            if (!m) {
                throw new Error(`invalid Gemfile.lock dependency line ${i}: ${JSON.stringify(line)}`)
            }
            const gemName = m[1]
            const gemVersion = m[2]

            const range = new sourcegraph.Range(
                new sourcegraph.Position(i, '    '.length),
                new sourcegraph.Position(i, line.length)
            )

            // A gem might be listed as an indirect dep before being listed as a direct dep.
            let gem = byName.get(gemName)
            if (gem) {
                gem.direct = true
                gem.range = range
                gem.version = gemVersion
            } else {
                gem = {
                    name: gemName,
                    version: gemVersion,
                    direct: true,
                    range,
                    directAncestors: [],
                }
                byName.set(gemName, gem)
            }
            parentDependency = gem
        } else if (line.startsWith('      ')) {
            // Transitive dependency.
            const m = line.match(/^\s*(.+?)(?: \((.+)\))?$/)
            if (!m) {
                throw new Error(`invalid Gemfile.lock transitive dependency line ${i}: ${JSON.stringify(line)}`)
            }
            const gemName = m[1]
            const versionTags = m[2] ? m[2].split(/,\s*/) : []
            // TODO!(sqs): take only 1st version
            const gemVersion = versionTags[0]

            if (!parentDependency) {
                throw new Error(
                    `invalid Gemfile.lock transitive dependency at line ${i} with no parent dependency: ${JSON.stringify(
                        line
                    )}`
                )
            }

            const existing = byName.get(gemName)
            if (existing) {
                existing.directAncestors.push(parentDependency.name)
            } else {
                const gem: RubyGemfileDependency = {
                    name: gemName,
                    version: gemVersion,
                    direct: false,
                    directAncestors: [parentDependency.name],
                }
                byName.set(gemName, gem)
            }
        }
    }
    return toDependencyList(byName)
}

const SUPPORT_MISSING_GEMFILE_LOCK = false

/**
 * Returns a list of all direct and indirect (i.e., all transitive) dependencies for a repository tree.
 */
export const rubyGemfileDependencies = async (
    files: { Gemfile: string; ['Gemfile.lock']?: string },
    context: BundlerExecutionContext
): Promise<RubyGemfileDependency[]> => {
    // Prefer using Gemfile.lock.
    if (files['Gemfile.lock']) {
        return parseRubyGemfileLock(files['Gemfile.lock'])
    }

    if (!SUPPORT_MISSING_GEMFILE_LOCK) {
        return []
    }

    const result = await executeBundlerCommand({
        commands: [['bundler', 'lock']],
        context,
    })
    return parseRubyGemfileLock(result.files['Gemfile.lock'])
    /* const guardEntry: RubyGemfileDependency = {
        name: 'guard',
        version: '>=1.2.0',
        range: new sourcegraph.Range(1, 2, 3, 4),
    }
    const omniauthEntry: RubyGemfileDependency = {
        name: 'omniauth',
        version: '>=2.9.1',
        range: new sourcegraph.Range(1, 2, 3, 4),
    }
    return [
        { name: 'rails', version: '=1.2.3', range: new sourcegraph.Range(1, 2, 3, 4) },
        { name: 'rake', version: '>=9.1.0', range: new sourcegraph.Range(1, 2, 3, 4) },
        guardEntry,
        { name: 'guard-rspec', version: '~5.3', directAncestors: [guardEntry] },
        omniauthEntry,
        { name: 'omniauth-openid', version: '=1.5.3', directAncestors: [omniauthEntry] },
    ] */
}
