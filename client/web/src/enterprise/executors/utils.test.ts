import { isExecutorVersionOutdated } from './utils'

interface TestExecutors {
    isActive: boolean

    executorVersion: string
    sourcegraphVersion: string

    isOutdated: boolean
}

describe('Executor Utils Test', () => {
    describe('isExecutorVersionOutdated', () => {
        const cases: TestExecutors[] = [
            // The executor isn't outdated when inactive.
            {
                executorVersion: '3.43.0',
                sourcegraphVersion: '3.42.0',
                isActive: false,
                isOutdated: false
            },
            // The executor isn't outdated when both sourcegraph and executor are the same (SEMVER).
            {
                executorVersion: '3.43.0',
                sourcegraphVersion: '3.43.0',
                isActive: true,
                isOutdated: false
            },
            // The executor isn't outdated when both sourcegraph and executor are the same (BuildDate).
            {
                executorVersion: 'executor-patch-notest-es-ignite-debug_168065_2022-08-25_e94e18c4ebcc_patch',
                sourcegraphVersion: '169135_2022-08-25_a2b623dce148',
                isActive: true,
                isOutdated: false
            },
            // The executor is outdated if the sourcegraph version is greater than theexecutor version (SEMVER).
            {
                executorVersion: '3.42.0',
                sourcegraphVersion: '3.43.0',
                isActive: true,
                isOutdated: true
            },
            // The executor is outdated if the sourcegraph version is greater than the executor version (BuildDate).
            {
                executorVersion: 'executor-patch-notest-es-ignite-debug_168065_2022-08-20_e94e18c4ebcc_patch',
                sourcegraphVersion: '169135_2022-08-25_a2b623dce148',
                isActive: true,
                isOutdated: true
            },
            // The executor is not outdated if the executor version is greater than the sourcegraph version (SEMVER)
            {
                executorVersion: '3.43.0',
                sourcegraphVersion: '3.42.0',
                isActive: true,
                isOutdated: false
            },
            // The executor is not outdated if the executor version is greater than the sourcegraph version (BuildDate)
            {
                executorVersion: 'executor-patch-notest-es-ignite-debug_168065_2022-08-20_e94e18c4ebcc_patch',
                sourcegraphVersion: '169135_2022-08-15_a2b623dce148',
                isActive: true,
                isOutdated: false
            },
        ]

        test.each(cases)(
            'Executor version $executorVersion with sourcegraphVersion $sourcegraphVersion, returns $isOutdated',
            ({ sourcegraphVersion, isActive, isOutdated, executorVersion }) => {
                expect(isExecutorVersionOutdated(isActive, executorVersion, sourcegraphVersion)).toBe(isOutdated)
            }
        )
    })
})
