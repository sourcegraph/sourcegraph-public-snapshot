import { NotificationType } from '@sourcegraph/extension-api-classes'
import { Diagnostic } from '@sourcegraph/extension-api-types'
import { Subscription, Unsubscribable } from 'rxjs'
import { map, mapTo } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'

export const ADD_DIAGNOSTIC_TO_THREAD_COMMAND: sourcegraph.Command = {
    command: 'sourcegraph.addDiagnosticToThread',
    title: 'Add diagnostic to thread',
}

export const addDiagnosticsToThread = (input: GQL.IAddDiagnosticsToThreadOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation AddDiagnosticsToThread($thread: ID!, $rawDiagnostics: [String!]!) {
                addDiagnosticsToThread(thread: $thread, rawDiagnostics: $rawDiagnostics) {
                    __typename
                }
            }
        `,
        input
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(undefined)
        )
        .toPromise()

/**
 * Registers contributions for thread diagnostics.
 */
export function register({ extensionsController }: ExtensionsControllerProps): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        extensionsController.services.commands.registerCommand({
            command: ADD_DIAGNOSTIC_TO_THREAD_COMMAND.command,
            run: async (diagnostic: Diagnostic) => {
                const thread = prompt('GraphQL ID of thread:', 'VGhyZWFkOjQ3OQ==') // id==479 TODO!(sqs)
                if (thread !== null) {
                    await addDiagnosticsToThread({ thread, rawDiagnostics: [JSON.stringify(diagnostic)] })
                    extensionsController.services.notifications.showMessages.next({
                        message: `Added diagnostic to thread`,
                        type: NotificationType.Info,
                    })
                }
            },
        })
    )
    return subscriptions
}
