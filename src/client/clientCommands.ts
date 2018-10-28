import { isArray } from 'lodash-es'
import { from, Subscription, throwError, Unsubscribable } from 'rxjs'
import { switchMap, take } from 'rxjs/operators'
import { Controller } from 'sourcegraph/module/client/controller'
import {
    ActionContributionClientCommandUpdateConfiguration,
    ConfigurationUpdateParams,
} from 'sourcegraph/module/protocol'
import { Context } from '../context'
import { isErrorLike } from '../errors'
import { ConfiguredExtension } from '../extensions/extension'
import { ConfigurationCascade, ConfigurationSubject, Settings } from '../settings'

/**
 * Registers the builtin client commands that are required for Sourcegraph extensions. See
 * {@link module:sourcegraph.module/protocol/contribution.ActionContribution#command} for
 * documentation.
 */
export function registerBuiltinClientCommands<S extends ConfigurationSubject, C extends Settings>(
    context: Pick<Context<S, C>, 'configurationCascade' | 'updateExtensionSettings' | 'queryGraphQL' | 'queryLSP'>,
    controller: Controller<ConfiguredExtension, ConfigurationCascade<S, C>>
): Unsubscribable {
    const subscription = new Subscription()

    subscription.add(
        controller.registries.commands.registerCommand({
            command: 'open',
            run: (url: string) => {
                // The `open` client command is usually implemented by ActionItem rendering the action with the
                // HTML <a> element, not by handling it here. Using an HTML <a> element means it is a standard
                // link, and native system behaviors such as open-in-new-tab work.
                //
                // If a client is not running in a web browser, this handler should be updated to call the system's
                // default URL handler using the system (e.g., Electron) API.
                window.open(url, '_blank')
                return Promise.resolve()
            },
        })
    )

    subscription.add(
        controller.registries.commands.registerCommand({
            command: 'openPanel',
            run: (viewID: string) => {
                // As above for `open`, the `openPanel` client command is usually implemented by an HTML <a>
                // element.
                window.open(urlForOpenPanel(viewID, window.location.hash))
                return Promise.resolve()
            },
        })
    )

    subscription.add(
        controller.registries.commands.registerCommand({
            command: 'updateConfiguration',
            run: (...anyArgs: any[]): Promise<void> => {
                const args = anyArgs as ActionContributionClientCommandUpdateConfiguration['commandArguments']
                return updateConfiguration(context, convertUpdateConfigurationCommandArgs(args))
            },
        })
    )

    /**
     * Sends a GraphQL request to the Sourcegraph GraphQL API and returns the result. The request is performed
     * with the privileges of the current user.
     */
    subscription.add(
        controller.registries.commands.registerCommand({
            command: 'queryGraphQL',
            run: (query: string, variables: { [name: string]: any }): Promise<any> =>
                // 🚨 SECURITY: The request might contain private info (such as
                // repository names), so the `mightContainPrivateInfo` parameter
                // is set to `true`. It is up to the client (e.g. browser
                // extension) to check that parameter and prevent the request
                // from being sent to Sourcegraph.com.
                from(context.queryGraphQL(query, variables, true)).toPromise(),
        })
    )

    /**
     * Sends a batched LSP request to the Sourcegraph LSP gateway API and returns the result. The request is
     * performed with the privileges of the current user.
     */
    subscription.add(
        controller.registries.commands.registerCommand({
            command: 'queryLSP',
            run: requests => from(context.queryLSP(requests)).toPromise(),
        })
    )

    return subscription
}

/**
 * Constructs the URL that will result in the panel being opened to the specified view. Other parameters in the URL
 * hash are preserved.
 *
 * @param viewID The ID of the view to open in the panel.
 * @param urlHash The current URL hash (beginning with '#' if non-empty).
 */
export function urlForOpenPanel(viewID: string, urlHash: string): string {
    // Preserve the existing URL fragment, if any.
    const params = new URLSearchParams(urlHash.slice('#'.length))
    params.set('tab', viewID)
    // In the URL fragment, the 'L1:2-3:4' is treated as a parameter with no value. Undo the escaping of ':'
    // and the addition of the '=' for the empty value, for aesthetic reasons.
    const paramsString = params
        .toString()
        .replace(/%3A/g, ':')
        .replace(/=&/g, '&')
    return `#${paramsString}`
}

/**
 * Applies an edit to the configuration settings of the highest-precedence subject.
 */
export function updateConfiguration<S extends ConfigurationSubject, C extends Settings>(
    context: Pick<Context<S, C>, 'configurationCascade' | 'updateExtensionSettings'>,
    params: ConfigurationUpdateParams
): Promise<void> {
    // TODO(sqs): Allow extensions to specify which subject's configuration to update
    // (instead of always updating the highest-precedence subject's configuration).
    return from(context.configurationCascade)
        .pipe(
            take(1),
            switchMap(x => {
                if (!x.subjects) {
                    return throwError(new Error('unable to update configuration: no configuration subjects available'))
                }
                if (isErrorLike(x.subjects)) {
                    return throwError(
                        new Error(
                            `unable to update configuration: error retrieving configuration subjects: ${
                                x.subjects.message
                            }`
                        )
                    )
                }
                const subject = x.subjects[x.subjects.length - 1].subject
                return context.updateExtensionSettings(subject.id, { edit: params })
            })
        )
        .toPromise()
}

/**
 * Converts the arguments for the `updateConfiguration` client command (as documented in
 * {@link ActionContributionClientCommandUpdateConfiguration#commandArguments})
 * to {@link ConfigurationUpdateParams}.
 */
export function convertUpdateConfigurationCommandArgs(
    args: ActionContributionClientCommandUpdateConfiguration['commandArguments']
): ConfigurationUpdateParams {
    if (
        !isArray(args) ||
        !(args.length >= 2 && args.length <= 4) ||
        !isArray(args[0]) ||
        !(args[2] === undefined || args[2] === null)
    ) {
        throw new Error(`invalid updateConfiguration arguments: ${JSON.stringify(args)}`)
    }
    const valueIsJSONEncoded = args.length === 4 && args[3] === 'json'
    return { path: args[0], value: valueIsJSONEncoded ? JSON.parse(args[1]) : args[1] }
}
