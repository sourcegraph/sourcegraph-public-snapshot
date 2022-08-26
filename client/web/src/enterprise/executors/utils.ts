import { parse, isAfter } from 'date-fns'
import semver from 'semver'

import { ExecutorFields } from '../../graphql-operations'

/**
 * Valid build date examples for sourcegraph
 * 169135_2022-08-25_a2b623dce148
 * 169120_2022-08-25_a94c7eb7beca
 *
 * Valid build date example for executor (patch)
 * executor-patch-notest-es-ignite-debug_168065_2022-08-18_e94e18c4ebcc_patch
 */
const buildDateRegex = /^[\w-]+_(\d{4}-\d{2}-\d{2})_\w+/
const developmentVersion = '0.0.0+dev'

export const isExecutorVersionOutdated = (
    isActive: ExecutorFields['active'],
    executorVersion: ExecutorFields['executorVersion'],
    sourcegraphVersion: string
): boolean => {
    const isDevelopment = executorVersion === developmentVersion && sourcegraphVersion === developmentVersion

    // Executors can only be outdated when they aren't in development mode and are active.
    if (!isDevelopment && isActive) {
        const semverExecutorVersion = semver.parse(executorVersion)
        const semverSourcegraphVersion = semver.parse(sourcegraphVersion)

        if (semverExecutorVersion && semverSourcegraphVersion) {
            // if the sourcegraph version is greater than the executor version, the
            // executor needs to be updated.
            return semver.gt(semverSourcegraphVersion, semverExecutorVersion)
        }

        // version is not in semver. We need to use the `buildDateRegex` to parse
        // the build date and compare.
        const sourcegraphBuildDateMatch = sourcegraphVersion.match(buildDateRegex)
        const executorBuildDateMatch = executorVersion.match(buildDateRegex)

        const isSourcegraphBuildDateValid = sourcegraphBuildDateMatch && sourcegraphBuildDateMatch.length > 1
        const isExecutorBuildDateValid = executorBuildDateMatch && executorBuildDateMatch.length > 1

        if (isSourcegraphBuildDateValid && isExecutorBuildDateValid) {
            const [, sourcegraphBuildDate] = sourcegraphBuildDateMatch
            const [, executorBuildDate] = executorBuildDateMatch

            /**
             * Syntax: isAfter(date, dateToCompare)
             *
             * date	            Date | Number	the date that should be after the other one to return true
             * dateToCompare	Date | Number	the date to compare with
             */
            return isAfter(
                parse(sourcegraphBuildDate, 'yyyy-MM-dd', new Date()),
                parse(executorBuildDate, 'yyyy-MM-dd', new Date())
            )
        }

        // if all of the above fail, we assume something is wrong.
        return true
    }

    return false
}
