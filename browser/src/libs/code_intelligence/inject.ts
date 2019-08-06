import { Observable, Subscription } from 'rxjs'
import { startWith } from 'rxjs/operators'
import { MutationRecordLike, observeMutations } from '../../shared/util/dom'
import { determineCodeHost, injectCodeIntelligenceToCodeHost } from './code_intelligence'

/**
 * Checks if the current page is a known code host. If it is,
 * injects features for the lifetime of the script in reaction to DOM mutations.
 *
 * @param isExtension `true` when executing in the browser extension.
 */
export async function injectCodeIntelligence(isExtension: boolean): Promise<Subscription> {
    const subscriptions = new Subscription()
    const codeHost = determineCodeHost()
    if (codeHost) {
        console.log('Detected code host:', codeHost.type)
        const mutations: Observable<MutationRecordLike[]> = observeMutations(document.body, {
            childList: true,
            subtree: true,
        }).pipe(startWith([{ addedNodes: [document.body], removedNodes: [] }]))
        subscriptions.add(await injectCodeIntelligenceToCodeHost(mutations, codeHost, isExtension))
    }
    return subscriptions
}
