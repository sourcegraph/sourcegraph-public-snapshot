import { Observable, Subscription } from 'rxjs'
import { filter, startWith } from 'rxjs/operators'

import { isCloudSupportedCodehost, SourcegraphUrlService } from '../../platform/sourcegraphUrlService'
import { MutationRecordLike, observeMutations as defaultObserveMutations } from '../../util/dom'

import { determineCodeHost, CodeHost, injectCodeIntelligenceToCodeHost, ObserveMutations } from './codeHost'
import { RepoIsBlockedForCloudError } from './errors'
import { logger } from './util/logger'

function inject(codeHost: CodeHost, assetsURL: string, sourcegraphURL: string, isExtension: boolean): Subscription {
    logger.info('Attaching code intelligence using', sourcegraphURL)

    const observeMutations: ObserveMutations = codeHost.observeMutations || defaultObserveMutations
    const mutations: Observable<MutationRecordLike[]> = observeMutations(document.body, {
        childList: true,
        subtree: true,
    }).pipe(startWith([{ addedNodes: [document.body], removedNodes: [] }]))

    return injectCodeIntelligenceToCodeHost(
        mutations,
        codeHost,
        {
            assetsURL,
            sourcegraphURL,
        },
        isExtension
    )
}

/**
 * Checks if the current page is a known code host. If it is,
 * injects features for the lifetime of the script in reaction to DOM mutations.
 *
 * @param isExtension `true` when executing in the browser extension.
 * @param onCodeHostFound setup logic to run when a code host is found (e.g. loading stylesheet) to avoid doing
 * such work on unsupported websites
 */
export async function injectCodeIntelligence(
    assetsURL: string,
    isExtension: boolean,
    onCodeHostFound?: (codeHost: CodeHost) => Promise<void>,
    overrideSourcegraphURL?: string
): Promise<Subscription> {
    const codeHost = determineCodeHost()
    if (!codeHost) {
        return new Subscription()
    }

    if (onCodeHostFound) {
        await onCodeHostFound(codeHost)
    }

    if (overrideSourcegraphURL) {
        return inject(codeHost, assetsURL, overrideSourcegraphURL, isExtension)
    }
    logger.info(`Detected codehost="${codeHost.type}"`)

    try {
        const { rawRepoName } = codeHost.getContext?.() || {}
        logger.info(`Detected repository="${rawRepoName ?? ''}"`)
        if (rawRepoName) {
            await SourcegraphUrlService.use(rawRepoName)
        }
    } catch (error) {
        if (error instanceof RepoIsBlockedForCloudError) {
            throw error
        }
    }

    return SourcegraphUrlService.observe(isExtension)
        .pipe(filter(isCloudSupportedCodehost))
        .subscribe(sourcegraphURL => inject(codeHost, assetsURL, sourcegraphURL, isExtension))
}
