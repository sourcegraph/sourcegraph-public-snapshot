import { Observable, of, BehaviorSubject } from 'rxjs'
import { distinctUntilChanged, filter, switchMap } from 'rxjs/operators'

import { observeStorageKey, setStorageKey, storage } from '../../../browser-extension/web-extension-api/storage'
import { SyncStorageItems } from '../../../browser-extension/web-extension-api/types'
import { RepoIsBlockedForCloudError } from '../../code-hosts/shared/errors'
import { CLOUD_SOURCEGRAPH_URL, isCloudSourcegraphUrl } from '../../util/context'

import { isInBlocklist } from './lib/isInBlocklist'
import { isRepoCloned } from './lib/isRepoCloned'

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
        use: async (rawRepoName: string): Promise<void> => {
            const selfHostedUrl = selfHostedSourcegraphURL.value

            // 1. repo is in blocklist for cloud
            if (isBlocked(rawRepoName, blocklist.value)) {
                if (selfHostedUrl) {
                    // 1.1. use self-hosted if it exist
                    currentSourcegraphURL.next(selfHostedUrl)
                } else {
                    // 1.2 throw error otherwise
                    throw new RepoIsBlockedForCloudError('Repository is in blocklist.')
                }
            } else if (await isRepoCloned(CLOUD_SOURCEGRAPH_URL, rawRepoName)) {
                // 3. repo exist in cloud
                currentSourcegraphURL.next(CLOUD_SOURCEGRAPH_URL)
            } else if (selfHostedUrl && (await isRepoCloned(selfHostedUrl, rawRepoName))) {
                // 4. repo exist in self-hosted
                currentSourcegraphURL.next(selfHostedUrl)
            } else {
                // 5. default use cloud
                currentSourcegraphURL.next(CLOUD_SOURCEGRAPH_URL)
            }
        },
        observeSelfHostedOrCloud: () =>
            selfHostedSourcegraphURL
                .asObservable()
                .pipe(
                    switchMap(selfHostedUrl => (selfHostedUrl ? of(selfHostedUrl) : SourcegraphUrlService.observe()))
                ),
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
