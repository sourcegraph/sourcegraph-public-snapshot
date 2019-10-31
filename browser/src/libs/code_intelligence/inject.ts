import { Observable, Subscription } from 'rxjs'
import { startWith } from 'rxjs/operators'
import { MutationRecordLike, observeMutations } from '../../shared/util/dom'
import { determineCodeHost, injectCodeIntelligenceToCodeHost } from './code_intelligence'
import { SourcegraphIntegrationURLs } from '../../platform/context'
import { checkUserLoggedIn } from '../../shared/backend/server'
import { requestGraphQLHelper } from '../../shared/backend/requestGraphQL'

/**
 * Checks if the current page is a known code host. If it is,
 * injects features for the lifetime of the script in reaction to DOM mutations.
 *
 * @param isExtension `true` when executing in the browser extension.
 */
export async function injectCodeIntelligence(
    urls: SourcegraphIntegrationURLs,
    isExtension: boolean
): Promise<Subscription> {
    const subscriptions = new Subscription()
    const codeHost = determineCodeHost()
    if (codeHost) {
        console.log('Detected code host:', codeHost.type)
        const userLoggedIn = await checkUserLoggedIn(requestGraphQLHelper(isExtension, urls.sourcegraphURL)).toPromise()
        // Exit early when the user is not logged in to the Sourcegraph instance.
        if (!userLoggedIn) {
            console.warn(`Sourcegraph is disabled: you must be logged in to ${urls.sourcegraphURL} to use Sourcegraph.`)
            return subscriptions
        }
        const mutations: Observable<MutationRecordLike[]> = observeMutations(document.body, {
            childList: true,
            subtree: true,
        }).pipe(startWith([{ addedNodes: [document.body], removedNodes: [] }]))
        subscriptions.add(injectCodeIntelligenceToCodeHost(mutations, codeHost, urls, isExtension, userLoggedIn))
    }
    return subscriptions
}
