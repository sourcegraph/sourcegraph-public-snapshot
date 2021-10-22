/* eslint-disable @typescript-eslint/ban-ts-comment */
import { BehaviorSubject } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'
import { TestScheduler } from 'rxjs/testing'
import sinon from 'sinon'

import { RepoIsBlockedForCloudError } from '../../code-hosts/shared/errors'
import { CLOUD_SOURCEGRAPH_URL } from '../../util/context'

import { createSourcegraphUrlService, SourcegraphUrlService as OriginalSourcegraphUrlService } from '.'

const scheduler = (): TestScheduler => new TestScheduler((a, b) => expect(a).toEqual(b))

const SELF_HOSTED_URL_STORAGE_KEY = 'sourcegraphURL'
const SELF_HOSTED_URL = 'https://self-hosted-sourcegraph.org'
const PUBLIC_REPO = 'PUBLIC_REPO'
const PRIVATE_REPO = 'PRIVATE_REPO'
const BLOCKED_REPO = 'BLOCKED_REPO'

type CreateMockParametersReturn = NonNullable<Parameters<typeof createSourcegraphUrlService>[0]>

const createMockParameters = (initialStorage = {}): CreateMockParametersReturn => {
    const mockStorage = new BehaviorSubject<Record<string, Record<string, any>>>(initialStorage)
    const isInBlocklist: CreateMockParametersReturn['isInBlocklist'] = (repoName: string) => repoName === BLOCKED_REPO
    const isStorageAvailable: CreateMockParametersReturn['isStorageAvailable'] = () => true

    // @ts-ignore
    const observeStorageKey: CreateMockParametersReturn['observeStorageKey'] = (areaName, key) =>
        mockStorage.asObservable().pipe(
            // @ts-ignore
            // eslint-disable-next-line @typescript-eslint/no-unsafe-return
            map(areas => areas[areaName]?.[key]),
            distinctUntilChanged()
        )

    const setStorageKey: CreateMockParametersReturn['setStorageKey'] = (areaName, key, value) =>
        new Promise<void>(resolve => {
            const newStorageValue = mockStorage.value
            newStorageValue[areaName] ??= {}

            // @ts-ignore
            // eslint-disable-next-line @typescript-eslint/no-unsafe-return
            newStorageValue[areaName][key] = value
            mockStorage.next(newStorageValue)
            resolve()
        })

    const isRepoCloned: CreateMockParametersReturn['isRepoCloned'] = async (url, repoName) =>
        Promise.resolve([PUBLIC_REPO, PRIVATE_REPO, BLOCKED_REPO].includes(repoName))

    return { isInBlocklist, isStorageAvailable, observeStorageKey, setStorageKey, isRepoCloned }
}

describe('SourcegraphUrlService', () => {
    let SourcegraphUrlService: typeof OriginalSourcegraphUrlService

    beforeEach(() => {
        SourcegraphUrlService = createSourcegraphUrlService(createMockParameters())
    })
    describe('.observe(false)', () => {
        afterEach(() => {
            delete window.SOURCEGRAPH_URL
            window.localStorage.removeItem('SOURCEGRAPH_URL')
        })

        it('returns window.SOURCEGRAPH_URL if it exists', () => {
            window.SOURCEGRAPH_URL = 'mock_url'
            scheduler().run(({ expectObservable }) => {
                expectObservable(SourcegraphUrlService.observe(false)).toBe('(0|)', [window.SOURCEGRAPH_URL])
            })
        })

        it('returns window.localStorage.SOURCEGRAPH_URL if it exists', () => {
            localStorage.setItem('SOURCEGRAPH_URL', 'local_storage_mock')
            scheduler().run(({ expectObservable }) => {
                expectObservable(SourcegraphUrlService.observe(false)).toBe('(0|)', [
                    localStorage.getItem('SOURCEGRAPH_URL'),
                ])
            })
        })

        it('returns cloud by default', () => {
            scheduler().run(({ expectObservable }) => {
                expectObservable(SourcegraphUrlService.observe(false)).toBe('(0|)', [CLOUD_SOURCEGRAPH_URL])
            })
        })
    })

    describe('self-hosted URL exists', () => {
        beforeEach(() => {
            SourcegraphUrlService = createSourcegraphUrlService(
                createMockParameters({
                    sync: { [SELF_HOSTED_URL_STORAGE_KEY]: SELF_HOSTED_URL },
                })
            )
        })

        describe('.observe(true)', () => {
            it('returns self-hosted URL if it exists', () => {
                scheduler().run(({ expectObservable }) => {
                    expectObservable(SourcegraphUrlService.observe()).toBe('0', [SELF_HOSTED_URL])
                })
            })

            it('returns cloud URL when self-hosted is removed', async () => {
                await SourcegraphUrlService.setSelfHostedURL(undefined)
                scheduler().run(({ expectObservable }) => {
                    expectObservable(SourcegraphUrlService.observe()).toBe('0', [CLOUD_SOURCEGRAPH_URL])
                })
            })
        })

        describe('.use + observe(true)', () => {
            it('returns self-hosted URL if repo exists there', async () => {
                await SourcegraphUrlService.use(PUBLIC_REPO)
                scheduler().run(({ expectObservable }) => {
                    expectObservable(SourcegraphUrlService.observe()).toBe('0', [SELF_HOSTED_URL])
                })
            })

            it('returns self-hosted URL if cloud is blocked', async () => {
                await SourcegraphUrlService.use(BLOCKED_REPO)
                scheduler().run(({ expectObservable }) => {
                    expectObservable(SourcegraphUrlService.observe()).toBe('0', [SELF_HOSTED_URL])
                })
            })

            it('returns cloud URL by default', async () => {
                await SourcegraphUrlService.use('some-non-existing-repo')
                scheduler().run(({ expectObservable }) => {
                    expectObservable(SourcegraphUrlService.observe()).toBe('0', [CLOUD_SOURCEGRAPH_URL])
                })
            })
        })

        describe('.observeSelfHostedURL', () => {
            it('returns self-hosted URL', () => {
                scheduler().run(({ expectObservable }) => {
                    expectObservable(SourcegraphUrlService.observeSelfHostedURL()).toBe('0', [SELF_HOSTED_URL])
                })
            })
        })
    })

    describe('self-hosted URL DOES NOT exists', () => {
        describe('.observe(true)', () => {
            it('returns cloud URL by default', () => {
                scheduler().run(({ expectObservable }) => {
                    expectObservable(SourcegraphUrlService.observe()).toBe('0', [CLOUD_SOURCEGRAPH_URL])
                })
            })

            it('returns cloud URL if self-hosted is empty', () => {
                SourcegraphUrlService = createSourcegraphUrlService(
                    createMockParameters({
                        sync: { [SELF_HOSTED_URL_STORAGE_KEY]: '' },
                    })
                )
                scheduler().run(({ expectObservable }) => {
                    expectObservable(SourcegraphUrlService.observe()).toBe('0', [CLOUD_SOURCEGRAPH_URL])
                })
            })
        })

        describe('.use + observe(true)', () => {
            it('throws error if cloud is blocked', async () => {
                let thrown: Error | false = false
                try {
                    await SourcegraphUrlService.use(BLOCKED_REPO)
                } catch (error) {
                    thrown = error
                }
                expect(thrown).toBeInstanceOf(RepoIsBlockedForCloudError)
            })

            it('returns cloud URL if repo exists there', async () => {
                await SourcegraphUrlService.use(PUBLIC_REPO)
                scheduler().run(({ expectObservable }) => {
                    expectObservable(SourcegraphUrlService.observe()).toBe('0', [CLOUD_SOURCEGRAPH_URL])
                })
            })
        })

        describe('.observeSelfHostedURL', () => {
            it('returns empty self-hosted URL', () => {
                scheduler().run(({ expectObservable }) => {
                    expectObservable(SourcegraphUrlService.observeSelfHostedURL()).toBe('0', [undefined])
                })
            })
        })
    })
})
