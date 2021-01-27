import { isEqual } from 'lodash'
import { from, Observable, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, filter, first, map, mapTo, publishReplay, refCount } from 'rxjs/operators'
import type { HoverAlert } from 'sourcegraph'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { PlatformContext } from '../../../../../shared/src/platform/context'
import { Settings } from '../../../../../shared/src/settings/settings'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { isDefined, isNot } from '../../../../../shared/src/util/types'
import { MutationRecordLike } from '../../util/dom'
import { CodeHost } from './codeHost'
import { trackViews } from './views'
import { userNeedsToSetupPrivateInstance } from './hoverAlerts'
import { Controller as ExtensionsController } from '../../../../../shared/src/extensions/controller'
import { syncSubscription } from '../../../../../shared/src/api/util'

const NATIVE_TOOLTIP_HIDDEN = 'native-tooltip--hidden'
const NATIVE_TOOLTIP_TYPE = 'nativeTooltips'

/**
 * Defines a native tooltip that is present on a page and exposes operations for manipulating it.
 */
export interface NativeTooltip {
    /** The native tooltip HTML element. */
    element: HTMLElement
}

export function handleNativeTooltips(
    mutations: Observable<MutationRecordLike[]>,
    nativeTooltipsEnabled: Observable<boolean>,
    { nativeTooltipResolvers, name, getContext }: Pick<CodeHost, 'nativeTooltipResolvers' | 'name' | 'getContext'>,
    sourcegraphURL: string
): { nativeTooltipsAlert: Observable<HoverAlert>; subscription: Unsubscribable } {
    const nativeTooltips = mutations.pipe(trackViews(nativeTooltipResolvers || []))
    const nativeTooltipsAlert = nativeTooltips.pipe(
        first(),
        mapTo({
            type: NATIVE_TOOLTIP_TYPE,
            summary: {
                kind: MarkupKind.Markdown,
                value: `<small>Sourcegraph has hidden ${name}'s native hover tooltips. You can toggle this at any time: to enable the native tooltips run "Code host: prefer non-Sourcegraph hover tooltips" from the command palette or set <code>{"codeHost.useNativeTooltips": true}</code> in your user settings.</small>`,
            },
        }),
        // If we can't provide the user hovers because it's private code, don't hide native tooltips.
        // Otherwise we would have to show the user two alerts at the same time.
        filter(() => !userNeedsToSetupPrivateInstance({ getContext }, sourcegraphURL)),
        publishReplay(1),
        refCount()
    )
    return {
        nativeTooltipsAlert,
        subscription: nativeTooltips.subscribe(({ element, subscriptions }) => {
            subscriptions.add(
                // This subscription is correctly handled through the view's `subscriptions`
                // eslint-disable-next-line rxjs/no-nested-subscribe
                nativeTooltipsEnabled.subscribe(enabled => {
                    // If we can't provide the user hovers because it's private code, don't hide native tooltips.
                    // Otherwise we would have to show the user two alerts at the same time.
                    element.classList.toggle(
                        NATIVE_TOOLTIP_HIDDEN,
                        !enabled && !userNeedsToSetupPrivateInstance({ getContext }, sourcegraphURL)
                    )
                })
            )
        }),
    }
}

export function nativeTooltipsEnabledFromSettings(settings: PlatformContext['settings']): Observable<boolean> {
    return from(settings).pipe(
        map(({ final }) => final),
        filter(isDefined),
        filter(isNot<ErrorLike | Settings, ErrorLike>(isErrorLike)),
        map(settings => !!settings['codeHost.useNativeTooltips']),
        distinctUntilChanged((a, b) => isEqual(a, b)),
        publishReplay(1),
        refCount()
    )
}

export function registerNativeTooltipContributions(
    extensionsController: Pick<ExtensionsController, 'extHostAPI'>
): Unsubscribable {
    return syncSubscription(
        extensionsController.extHostAPI.then(extensionHostAPI =>
            extensionHostAPI.registerContributions({
                actions: [
                    {
                        id: 'codeHost.toggleUseNativeTooltips',
                        command: 'updateConfiguration',
                        category: 'Code host',
                        commandArguments: [
                            'codeHost.useNativeTooltips',
                            /* eslint-disable-next-line no-template-curly-in-string */
                            '${!config.codeHost.useNativeTooltips}',
                            null,
                            'json',
                        ],
                        title:
                            /* eslint-disable-next-line no-template-curly-in-string */
                            'Prefer ${config.codeHost.useNativeTooltips && "Sourcegraph" || "non-Sourcegraph"} hover tooltips',
                    },
                ],
                menus: {
                    commandPalette: [
                        {
                            action: 'codeHost.toggleUseNativeTooltips',
                        },
                    ],
                },
            })
        )
    )
}
