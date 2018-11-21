import { Subscribable } from 'rxjs'
import { MessageTransports } from '../api/protocol/jsonrpc2/connection'
import { ConfiguredExtension } from '../extensions/extension'
import { GraphQLResult } from '../graphql/graphql'
import * as GQL from '../graphql/schema'
import { UpdateExtensionSettingsArgs } from '../settings/edit'
import { SettingsCascade, SettingsCascadeOrError } from '../settings/settings'

/**
 * Platform-specific data and methods shared by multiple Sourcegraph components.
 *
 * Whenever shared code (in shared/) needs to perform an action or retrieve data that requires different
 * implementations depending on the platform, the shared code should use this value's fields.
 */
export interface PlatformContext {
    /**
     * An observable that emits whenever the settings cascade changes (including when any individual subject's
     * settings change).
     */
    readonly settingsCascade: Subscribable<SettingsCascadeOrError>

    /**
     * Update the settings for the subject.
     */
    updateSettings(subject: GQL.ID, args: UpdateExtensionSettingsArgs): Promise<void>

    /**
     * Sends a request to the Sourcegraph GraphQL API and returns the response.
     *
     * @template R The GraphQL result type
     * @param request The GraphQL request (query or mutation)
     * @param variables An object whose properties are GraphQL query name-value variable pairs
     * @param mightContainPrivateInfo 🚨 SECURITY: Whether or not sending the GraphQL request to Sourcegraph.com
     * could leak private information such as repository names.
     * @return Observable that emits the result or an error if the HTTP request failed
     */
    queryGraphQL<R extends GQL.IQuery | GQL.IMutation>(
        request: string,
        variables?: { [name: string]: any },
        mightContainPrivateInfo?: boolean
    ): Subscribable<GraphQLResult<R>>

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
     * Forces the currently displayed tooltip, if any, to update its contents.
     */
    forceUpdateTooltip(): void

    /**
     * Spawns (e.g., in a Web Worker) and opens a communication channel to an extension.
     */
    createMessageTransports: (
        extension: ConfiguredExtension,
        settingsCascade: SettingsCascade
    ) => Promise<MessageTransports>
}

/**
 * React partial props for components needing the {@link PlatformContext}.
 */
export interface PlatformContextProps {
    platformContext: PlatformContext
}
