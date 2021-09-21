import { Observable, of, from, merge, BehaviorSubject } from 'rxjs'
import { map, first, defaultIfEmpty, distinctUntilChanged, tap } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import { background } from '../../../browser-extension/web-extension-api/runtime'
import { observeStorageKey, storage } from '../../../browser-extension/web-extension-api/storage'
import { SyncStorageItems } from '../../../browser-extension/web-extension-api/types'
import { CLOUD_SOURCEGRAPH_URL, isCloudSourcegraphUrl } from '../../util/context'

import { isInBlocklist } from './lib/isInBlocklist'

const QUERY = gql`
    query ResolveRawRepoName($repoName: String!) {
        repository(name: $repoName) {
            mirrorInfo {
                cloned
            }
        }
    }
`
const isRepoCloned = (sourcegraphURL: string, repoName: string): Observable<boolean> =>
    from(
        background.requestGraphQL<GQL.IQuery>({
            request: QUERY,
            variables: { repoName },
            sourcegraphURL,
        })
    ).pipe(
        map(dataOrThrowErrors),
        map(({ repository }) => !!repository?.mirrorInfo?.cloned)
    )

export const SourcegraphUrlService = (() => {
    const selfHostedSourcegraphURL = new BehaviorSubject<string | undefined>(undefined)
    const currentSourcegraphURL = new BehaviorSubject<string>(CLOUD_SOURCEGRAPH_URL)
    const blocklist = new BehaviorSubject<SyncStorageItems['blocklist'] | undefined>(undefined)

    if (storage?.sync) {
        // eslint-disable-next-line rxjs/no-ignored-subscription
        observeStorageKey('sync', 'sourcegraphURL').subscribe(selfHostedSourcegraphURL)
        // eslint-disable-next-line rxjs/no-ignored-subscription
        observeStorageKey('sync', 'blocklist').subscribe(blocklist)
    }

    /* Checks if rawRepoName is blocked */
    const isBlocked = (sgURL: string, rawRepoName: string): boolean => {
        const { enabled = false, content = '' } = blocklist.value ?? {}

        return isCloudSourcegraphUrl(sgURL) && enabled && isInBlocklist(content, rawRepoName)
    }

    /**
     * Determines sourcegraph instance URL where a given rawRepoName exists.
     * Uses cache as well as network requests
     */
    const determineSourcegraphURL = async (rawRepoName: string): Promise<string | undefined> => {
        const { cache = {} } = await storage.sync.get('cache')

        const URLs = [CLOUD_SOURCEGRAPH_URL, selfHostedSourcegraphURL.value].filter(Boolean) as string[]

        const cachedURL = cache[rawRepoName]
        if (cachedURL && URLs.includes(cachedURL) && !isBlocked(cachedURL, rawRepoName)) {
            return cachedURL
        }

        return merge(
            ...URLs.filter(url => !isBlocked(url, rawRepoName)).map(url =>
                isRepoCloned(url, rawRepoName).pipe(map(isCloned => [isCloned, url] as [boolean, string]))
            )
        )
            .pipe(
                first(([isCloned]) => isCloned),
                map(([, url]) => url),
                defaultIfEmpty<string | undefined>(undefined),
                tap(url => {
                    if (url) {
                        cache[rawRepoName] = url
                        storage.sync.set({ cache }).catch(console.error)
                    }
                })
            )
            .toPromise()
    }

    return {
        /*  Returns currently used Sourcegraph URL */
        observe: (isExtension: boolean = true): Observable<string> => {
            if (!isExtension) {
                return of(
                    window.SOURCEGRAPH_URL || window.localStorage.getItem('SOURCEGRAPH_URL') || CLOUD_SOURCEGRAPH_URL
                )
            }

            return currentSourcegraphURL.asObservable().pipe(distinctUntilChanged())
        },
        /* Updates current used Sourcegraph URL based on the current rawRepoName */
        use: async (rawRepoName: string): Promise<void> => {
            const errorMessage = `Couldn't detect sourcegraphURL for the ${rawRepoName}\n`
            try {
                const sourcegraphURL = await determineSourcegraphURL(rawRepoName)
                if (!sourcegraphURL) {
                    throw new Error(errorMessage)
                }
                currentSourcegraphURL.next(sourcegraphURL)
            } catch (error) {
                // eslint-disable-next-line @typescript-eslint/restrict-plus-operands
                throw new Error(errorMessage + error?.toString())
            }
        },
        /* Get self-hosted Sourcegraph URL */
        getSelfHostedSourcegraphURL: () => selfHostedSourcegraphURL.asObservable(),
        /** Set self-hosted Sourcegraph URL */
        setSelfHostedSourcegraphURL: (sourcegraphURL?: string): Promise<void> => storage.sync.set({ sourcegraphURL }),
        getBlocklist: () => blocklist.asObservable(),
        setBlocklist: (blocklist: SyncStorageItems['blocklist']): Promise<void> => storage.sync.set({ blocklist }),
    }
})()
