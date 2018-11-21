import { throwError } from 'rxjs'
import { mergeMap, take } from 'rxjs/operators'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { mutateSettings, updateSettings } from '../../../../shared/src/settings/edit'
import storage from '../browser/storage'
import { getContext } from '../shared/backend/context'
import { requestGraphQL } from '../shared/backend/graphql'
import { sendLSPHTTPRequests } from '../shared/backend/lsp'
import { canFetchForURL, sourcegraphUrl } from '../shared/util/context'
import { createMessageTransports } from './messageTransports'
import { editClientSettings, settingsCascade, settingsCascadeRefreshes } from './settings'

/**
 * Creates the {@link PlatformContext} for the browser extension.
 */
export function createPlatformContext(): PlatformContext {
    // TODO: support listening for changes to sourcegraphUrl
    const sourcegraphLanguageServerURL = new URL(sourcegraphUrl)
    sourcegraphLanguageServerURL.pathname = '.api/xlang'

    const context: PlatformContext = {
        settingsCascade,
        updateSettings: async (subject, args) => {
            try {
                await updateSettings(
                    context,
                    subject,
                    args,
                    // Support storing settings on the client (in the browser extension) so that unauthenticated
                    // Sourcegraph viewers can update settings.
                    subject === 'Client' ? () => editClientSettings(args) : mutateSettings
                )
            } finally {
                settingsCascadeRefreshes.next()
            }
        },
        queryGraphQL: (request, variables, requestMightContainPrivateInfo) =>
            storage.observeSync('sourcegraphURL').pipe(
                take(1),
                mergeMap(url =>
                    requestGraphQL({
                        ctx: getContext({ repoKey: '', isRepoSpecific: false }),
                        request,
                        variables,
                        url,
                        requestMightContainPrivateInfo,
                    })
                )
            ),
        queryLSP: canFetchForURL(sourcegraphUrl)
            ? requests => sendLSPHTTPRequests(requests)
            : () =>
                  throwError(
                      'The queryLSP command is unavailable because the current repository does not exist on the Sourcegraph instance.'
                  ),
        forceUpdateTooltip: () => {
            // TODO(sqs): implement tooltips on the browser extension
        },
        createMessageTransports,
    }
    return context
}
