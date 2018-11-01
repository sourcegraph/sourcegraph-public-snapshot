import { Subscribable } from 'rxjs'
import { ConfigurationUpdateParams } from 'sourcegraph/module/protocol'
import { Controller } from './controller'
import { QueryResult } from './graphql'
import * as GQL from './schema/graphqlschema'
import { ConfigurationCascadeOrError, ConfigurationSubject, ID, Settings } from './settings'

export type UpdateExtensionSettingsArgs =
    | { edit?: ConfigurationUpdateParams }
    | {
          extensionID: string
          // TODO: unclean api, allows 4 states (2 bools), but only 3 are valid (none/disabled/enabled)
          enabled?: boolean
          remove?: boolean
      }

/**
 * Description of the context in which extensions-client-common is running, and platform-specific hooks.
 */
export interface Context<S extends ConfigurationSubject, C extends Settings> {
    /**
     * An observable that emits whenever the configuration cascade changes (including when any individual subject's
     * settings change).
     */
    readonly configurationCascade: Subscribable<ConfigurationCascadeOrError<S, C>>

    updateExtensionSettings(subject: ID, args: UpdateExtensionSettingsArgs): Subscribable<void>

    /**
     * Sends a request to the Sourcegraph GraphQL API and returns the response.
     *
     * @param request The GraphQL request (query or mutation)
     * @param variables An object whose properties are GraphQL query name-value variable pairs
     * @param mightContainPrivateInfo 🚨 SECURITY: Whether or not sending the GraphQL request to Sourcegraph.com
     * could leak private information such as repository names.
     * @return Observable that emits the result or an error if the HTTP request failed
     */
    queryGraphQL(
        request: string,
        variables?: { [name: string]: any },
        mightContainPrivateInfo?: boolean
    ): Subscribable<QueryResult<Pick<GQL.IQuery, 'extensionRegistry'>>>

    /**
     * Sends a batch of LSP requests to the Sourcegraph LSP gateway API and returns the result.
     *
     * @param requests An array of LSP requests (with methods `initialize`, the (optional) request, `shutdown`,
     *                 `exit`).
     * @return Observable that emits the result and then completes, or an error if the request fails. The value is
     *         an array of LSP responses.
     */
    queryLSP(requests: object[]): Subscribable<object[]>

    /**
     * React components for icons. They are expected to size themselves appropriately with the surrounding DOM flow
     * content.
     */
    readonly icons: Record<
        'Loader' | 'Warning' | 'Info' | 'Menu' | 'CaretDown' | 'Add' | 'Settings',
        React.ComponentType<{
            className: 'icon-inline' | string
            onClick?: () => void
        }>
    >

    /**
     * Forces the currently displayed tooltip, if any, to update its contents.
     */
    forceUpdateTooltip(): void
}

/**
 * React partial props for components needing the extensions controller.
 */
export interface ExtensionsProps<S extends ConfigurationSubject, C extends Settings> {
    extensions: Controller<S, C>
}
