import { ConfigurationUpdateParams } from 'cxp/module/protocol'
import { Observable } from 'rxjs'
import { Controller } from './controller'
import { Settings } from './copypasta'
import { QueryResult } from './graphql'
import * as GQL from './schema/graphqlschema'
import { ConfigurationCascade, ConfigurationSubject, ID } from './settings'

/**
 * Description of the context in which extensions-client-common is running, and platform-specific hooks.
 */
export interface Context<S extends ConfigurationSubject, C = Settings> {
    /**
     * An observable that emits whenever the configuration cascade changes (including when any individual subject's
     * settings change).
     */
    readonly configurationCascade: Observable<ConfigurationCascade<S, C>>

    /**
     * Updates the extension settings for extensionID and for the given subject.
     */
    updateExtensionSettings(
        subject: ID,
        args: {
            extensionID: string
            edit?: ConfigurationUpdateParams
            // TODO: unclean api, allows 4 states (2 bools), but only 3 are valid (none/disabled/enabled)
            enabled?: boolean
            remove?: boolean
        }
    ): Observable<void>

    /**
     * Sends a request to the Sourcegraph GraphQL API and returns the response.
     *
     * @param request The GraphQL request (query or mutation)
     * @param variables An object whose properties are GraphQL query name-value variable pairs
     * @return Observable that emits the result or an error if the HTTP request failed
     */
    queryGraphQL(
        request: string,
        variables?: { [name: string]: any }
    ): Observable<QueryResult<Pick<GQL.IQuery, 'extensionRegistry'>>>

    /**
     * React components for icons. They are expected to size themselves appropriately with the surrounding DOM flow
     * content.
     */
    readonly icons: Record<
        'Loader' | 'Warning' | 'Menu' | 'CaretDown',
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
export interface ExtensionsProps<S extends ConfigurationSubject, C = Settings> {
    extensions: Controller<S, C>
}
