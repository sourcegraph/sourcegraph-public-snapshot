import { last } from 'lodash'
import { Observable, of, BehaviorSubject } from 'rxjs'
import { distinctUntilChanged, filter, map } from 'rxjs/operators'

import {
    isStorageAvailable as originalIsStorageAvailable,
    observeStorageKey as originalObserveStorageKey,
    setStorageKey as originalSetStorageKey,
} from '../../../browser-extension/web-extension-api/storage'
import { SyncStorageItems } from '../../../browser-extension/web-extension-api/types'
import { RepoIsBlockedForCloudError } from '../../code-hosts/shared/errors'
import { CLOUD_SOURCEGRAPH_URL, isCloudSourcegraphUrl } from '../../util/context'

import { isInBlocklist as originalIsInBlocklist } from './lib/isInBlocklist'
import { isRepoCloned as originalIsRepoCloned } from './lib/isRepoCloned'

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
 * @description exporting for unit testing purposes only
 */
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
export const createSourcegraphUrlService = ({
    isInBlocklist = originalIsInBlocklist,
    setStorageKey = originalSetStorageKey,
    observeStorageKey = originalObserveStorageKey,
    isStorageAvailable = originalIsStorageAvailable,
    isRepoCloned = originalIsRepoCloned,
} = {}) => {
    const selfHostedURL = new BehaviorSubject<string | undefined>(undefined)
    const blocklist = new BehaviorSubject<SyncStorageItems['blocklist'] | undefined>(undefined)
    const currentSourcegraphURL = new BehaviorSubject<string | undefined>(undefined)

    if (isStorageAvailable()) {
        observeStorageKey(STORAGE_AREA, 'sourcegraphURL')
            // filter since cloud url is already included
            .pipe(filter(sgURL => sgURL !== CLOUD_SOURCEGRAPH_URL))
            // eslint-disable-next-line rxjs/no-ignored-subscription
            .subscribe(selfHostedURL)

        observeStorageKey(STORAGE_AREA, 'blocklist')
            // eslint-disable-next-line rxjs/no-ignored-subscription
            .subscribe(blocklist)
    }

    /**
     * Observe sourcegraphURL
     */
    const observe = (isExtension: boolean = true): Observable<string> => {
        if (!isExtension) {
            return of(window.SOURCEGRAPH_URL || window.localStorage.getItem('SOURCEGRAPH_URL') || CLOUD_SOURCEGRAPH_URL)
        }

        return currentSourcegraphURL.asObservable().pipe(
            map(currentUrl => currentUrl || selfHostedURL.value || CLOUD_SOURCEGRAPH_URL),
            distinctUntilChanged()
        )
    }

    /**
     * Updates current sourcegraphURL based on the rawRepoName, self-hosted URL, blocklist
     */
    const use = async (rawRepoName: string): Promise<void> => {
        const instanceURLs = [
            selfHostedURL.value,
            ...(isInBlocklist(rawRepoName, blocklist.value) ? [] : [CLOUD_SOURCEGRAPH_URL]),
        ].filter(Boolean) as string[]

        let detectedURL: string | undefined

        for (const instanceURL of instanceURLs) {
            if (await isRepoCloned(instanceURL, rawRepoName)) {
                detectedURL = instanceURL
                break
            }
        }

        detectedURL ??= last(instanceURLs)
        if (!detectedURL) {
            throw new RepoIsBlockedForCloudError('Repository is in blocklist.')
        }
        currentSourcegraphURL.next(detectedURL)
    }

    /**
     * Set self-hosted Sourcegraph URL
     */
    const setSelfHostedURL = (sourcegraphURL?: string): Promise<void> =>
        setStorageKey(STORAGE_AREA, 'sourcegraphURL', sourcegraphURL)

    /**
     * Save blocklist to storage
     */
    const setBlocklist = async (blocklist: SyncStorageItems['blocklist']): Promise<void> =>
        setStorageKey(STORAGE_AREA, 'blocklist', blocklist)

    /**
     * Observe self-hosted Sourcegraph URL
     */
    const observeSelfHostedURL = (): Observable<string | undefined> => selfHostedURL.asObservable()

    /**
     * Get blocklist from storage
     */
    const observeBlocklist = (): Observable<SyncStorageItems['blocklist'] | undefined> => blocklist.asObservable()

    return {
        observe,
        observeBlocklist,
        observeSelfHostedURL,

        use,

        setBlocklist,
        setSelfHostedURL,
    }
}

export const SourcegraphUrlService = createSourcegraphUrlService()
