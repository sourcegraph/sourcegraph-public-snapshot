import { Observable, of, BehaviorSubject } from 'rxjs'
import { distinctUntilChanged, filter } from 'rxjs/operators'

import {
    observeStorageKey,
    setStorageKey,
    getStorageKey,
    storage,
} from '../../../browser-extension/web-extension-api/storage'
import { SyncStorageItems } from '../../../browser-extension/web-extension-api/types'
import { RepoIsBlockedForCloudError } from '../../code-hosts/shared/errors'
import { logger } from '../../code-hosts/shared/util/logger'
import { CLOUD_SOURCEGRAPH_URL, isCloudSourcegraphUrl } from '../../util/context'

import { firstURLWhereRepoExists } from './lib/firstUrlWhereRepoExists'
import { isInBlocklist } from './lib/isInBlocklist'

const CLOUD_SUPPORTED_CODE_HOST_HOSTS = ['github.com', 'gitlab.com']
const STORAGE_AREA = 'sync'

/**
 * Prevent repo lookups for code hosts that we know cannot have repositories
 * cloned on sourcegraph.com. Repo lookups trigger cloning, which will
 * inevitably fail in this case.
 */
export const isCloudSupportedCodehost = (sourcegraphURL: string): boolean => {
    if (!isCloudSourcegraphUrl(sourcegraphURL)) {
        return true
    }
    const { hostname } = new URL(location.href)
    if (CLOUD_SUPPORTED_CODE_HOST_HOSTS.some(cloudHost => cloudHost === hostname)) {
        return true
    }
    return false
}

/**
 * Checks if rawRepoName is blocked
 */
const isBlocked = (rawRepoName: string, blocklist?: SyncStorageItems['blocklist']): boolean => {
    const { enabled = false, content = '' } = blocklist ?? {}

    return enabled && isInBlocklist(content, rawRepoName)
}

// todo: add tests
export const SourcegraphUrlService = (() => {
    const selfHostedSourcegraphURL = new BehaviorSubject<string | undefined>(undefined)
    const currentSourcegraphURL = new BehaviorSubject<string>(CLOUD_SOURCEGRAPH_URL)
    const blocklist = new BehaviorSubject<SyncStorageItems['blocklist'] | undefined>(undefined)

    if (storage?.sync) {
        observeStorageKey(STORAGE_AREA, 'sourcegraphURL')
            // filter since cloud url is already included
            .pipe(filter(sgURL => sgURL !== CLOUD_SOURCEGRAPH_URL))
            // eslint-disable-next-line rxjs/no-ignored-subscription
            .subscribe(selfHostedSourcegraphURL)
        // eslint-disable-next-line rxjs/no-ignored-subscription
        observeStorageKey(STORAGE_AREA, 'blocklist').subscribe(blocklist)
    }

    return {
        /**
         * Observe sourcegraphURL
         *
         * @returns sourcegraphURL
         */
        observe: (isExtension: boolean = true): Observable<string> => {
            if (!isExtension) {
                return of(
                    window.SOURCEGRAPH_URL || window.localStorage.getItem('SOURCEGRAPH_URL') || CLOUD_SOURCEGRAPH_URL
                )
            }

            return currentSourcegraphURL.asObservable().pipe(distinctUntilChanged())
        },
        /**
         * Updates sourcegraphURL to use based on the rawRepoName, blocklist, self-hosted URL
         */
        use: async (rawRepoName: string, useCache: boolean = false): Promise<void> => {
            const isCloudBlocked = isBlocked(rawRepoName, blocklist.value)
            const cache = (await getStorageKey(STORAGE_AREA, 'cache')) || {}
            const selfHostedUrl = selfHostedSourcegraphURL.value
            const URLs = [
                ...(!isCloudBlocked && isCloudSupportedCodehost(CLOUD_SOURCEGRAPH_URL) ? [CLOUD_SOURCEGRAPH_URL] : []),
                selfHostedUrl,
            ].filter(Boolean) as string[]

            let detectedURL: string | undefined
            if (useCache && cache[rawRepoName] && URLs.includes(cache[rawRepoName] as string)) {
                detectedURL = cache[rawRepoName]
            } else {
                detectedURL = await firstURLWhereRepoExists(URLs, rawRepoName)
            }

            if (detectedURL) {
                logger.info('Detected sourcegraph', detectedURL)
                currentSourcegraphURL.next(detectedURL)
                cache[rawRepoName] = detectedURL
                setStorageKey(STORAGE_AREA, 'cache', cache).catch(console.error)
            } else if (isCloudBlocked && selfHostedUrl) {
                logger.info('Repo is blocked. Falling back to self-hosted URL', detectedURL)
                currentSourcegraphURL.next(selfHostedUrl)
            } else if (isCloudBlocked) {
                throw new RepoIsBlockedForCloudError('Repository is in blocklist.')
            }
        },
        /**
         * Get self-hosted Sourcegraph URL
         */
        getSelfHostedSourcegraphURL: () => selfHostedSourcegraphURL.asObservable(),
        /**
         * Set self-hosted Sourcegraph URL
         */
        setSelfHostedSourcegraphURL: (sourcegraphURL?: string): Promise<void> =>
            setStorageKey(STORAGE_AREA, 'sourcegraphURL', sourcegraphURL),
        /**
         * Get blocklist from storage
         */
        getBlocklist: () => blocklist.asObservable(),
        /**
         * Save blocklist to storage
         */
        setBlocklist: (blocklist: SyncStorageItems['blocklist']): Promise<void> =>
            setStorageKey(STORAGE_AREA, 'blocklist', blocklist),
    }
})()
