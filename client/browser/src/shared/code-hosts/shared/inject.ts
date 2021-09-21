import { Observable, Subscription } from 'rxjs'
import { startWith, filter } from 'rxjs/operators'

import { SourcegraphUrlService } from '../../platform/sourcegraphUrlService'
import { CLOUD_SOURCEGRAPH_URL } from '../../util/context'
import { MutationRecordLike, observeMutations as defaultObserveMutations } from '../../util/dom'

import { determineCodeHost, CodeHost, injectCodeIntelligenceToCodeHost, ObserveMutations } from './codeHost'
import { logger } from './util/logger'

const CLOUD_SUPPORTED_CODE_HOST_HOSTS = ['github.com', 'gitlab.com']

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

    const { rawRepoName } = codeHost.getContext?.() || {}
    logger.info(`Detected: codehost="${codeHost.type}" repository="${rawRepoName ?? ''}"`)

    if (rawRepoName) {
        await SourcegraphUrlService.use(rawRepoName)
    }

    return SourcegraphUrlService.observe(isExtension)
        .pipe(
            filter(sourcegraphURL => {
                /*
                    /* Prevent repo lookups for code hosts that we know cannot have repositories
                    /* cloned on sourcegraph.com. Repo lookups trigger cloning, which will
                    /* inevitably fail in this case.
                    */
                if (sourcegraphURL !== CLOUD_SOURCEGRAPH_URL) {
                    return true
                }
                const { hostname } = new URL(location.href)
                if (CLOUD_SUPPORTED_CODE_HOST_HOSTS.some(cloudHost => cloudHost === hostname)) {
                    return true
                }
                console.error(
                    `Sourcegraph code host integration: stopped initialization since ${hostname} is not a supported code host when Sourcegraph URL is ${CLOUD_SOURCEGRAPH_URL}.\n List of supported code hosts on ${CLOUD_SOURCEGRAPH_URL}: ${CLOUD_SUPPORTED_CODE_HOST_HOSTS.join(
                        ', '
                    )}`
                )
                return false
            })
        )
        .subscribe(sourcegraphURL => inject(codeHost, assetsURL, sourcegraphURL, isExtension))
}
