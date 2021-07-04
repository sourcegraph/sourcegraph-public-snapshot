import { Subscription, Unsubscribable } from 'rxjs'
import { map } from 'rxjs/operators'

import { syncRemoteSubscription } from '@sourcegraph/shared/src/api/util'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { ENTERPRISE_HOVER_ACTIONS_CONTEXT } from '@sourcegraph/shared/src/hover/actions'
import { parseRepoURI } from '@sourcegraph/shared/src/util/url'

import { getExploreUsageURL } from '@sourcegraph/web/src/enterprise/guide/getExploreUsageUrl'

ENTERPRISE_HOVER_ACTIONS_CONTEXT.push(parameters => {
    const { repoName, revision, commitID, filePath } = parseRepoURI(parameters.textDocument.uri)
    return getExploreUsageURL({
        repo: repoName,
        commitID: revision || commitID!,
        path: filePath || '',
        line: parameters.position.line,
        character: parameters.position.character,
    }).pipe(map(exploreUsageURL => ({ 'exploreUsage.url': exploreUsageURL })))
})

export const registerExploreUsageActionContribution = ({
    extensionsController,
}: ExtensionsControllerProps<'extHostAPI'>): Unsubscribable => {
    const subscriptions = new Subscription()

    subscriptions.add(
        syncRemoteSubscription(
            extensionsController.extHostAPI.then(extensionHostAPI =>
                extensionHostAPI.registerContributions({
                    actions: [
                        {
                            id: 'exploreUsage',
                            title: 'Explore usage',
                            command: 'open',
                            // eslint-disable-next-line no-template-curly-in-string
                            commandArguments: ['${exploreUsage.url}'],
                        },
                    ],
                    menus: {
                        hover: [{ action: 'exploreUsage', when: 'exploreUsage.url && goToDefinition.url' }],
                    },
                })
            )
        )
    )

    return subscriptions
}
