import { npmPackageManager } from './npm/npm'
import { yarnPackageManager } from './yarn/yarn'
import { combinedProvider } from '../dependencyManagement/combinedProvider'

const PROVIDERS = [npmPackageManager, yarnPackageManager]

export const packageJsonDependencyManagementProviderRegistry = combinedProvider(PROVIDERS)

/*
                    let matchRange = findMatchRange(hit.packageJson.text!, `"${packageName}"`)
                    let matchDoc: sourcegraph.TextDocument | undefined
                    if (matchRange) {
                        matchDoc = hit.packageJson
                    }
                    if (!matchRange) {
                        matchRange = findMatchRange(
                            hit.lockfile.text!,
                            type === 'npm' ? `"${packageName}"` : `${packageName}@`
                        )
                        if (matchRange) {
                            matchDoc = hit.lockfile
                        }
                    }

                    if (!matchRange || !matchDoc) {
                        return null
                    }


										------

										function findMatchRange(text: string, str: string): sourcegraph.Range | null {
    for (const [i, line] of text.split('\n').entries()) {
        const j = line.indexOf(str)
        if (j !== -1) {
            return new sourcegraph.Range(i, j, i, j + str.length)
        }
    }
    return null
}

										*/
