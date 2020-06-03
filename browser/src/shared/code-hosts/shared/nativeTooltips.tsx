import { isEqual } from 'lodash'
import * as React from 'react'
import { from, Observable, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, filter, first, map, mapTo, publishReplay, refCount } from 'rxjs/operators'
import { parseTemplate } from '../../../../../shared/src/api/client/context/expr/evaluator'
import { Services } from '../../../../../shared/src/api/client/services'
import { HoverAlert } from '../../../../../shared/src/hover/HoverOverlay'
import { PlatformContext } from '../../../../../shared/src/platform/context'
import { Settings } from '../../../../../shared/src/settings/settings'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { isDefined, isNot } from '../../../../../shared/src/util/types'
import { MutationRecordLike } from '../../util/dom'
import { CodeHost } from './codeHost'
import { ExtensionHoverAlertType } from './hoverAlerts'
import { trackViews } from './views'

const NATIVE_TOOLTIP_HIDDEN = 'native-tooltip--hidden'

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
    { nativeTooltipResolvers, name }: Pick<CodeHost, 'nativeTooltipResolvers' | 'name'>
): { nativeTooltipsAlert: Observable<HoverAlert<ExtensionHoverAlertType>>; subscription: Unsubscribable } {
    const nativeTooltips = mutations.pipe(trackViews(nativeTooltipResolvers || []))
    const nativeTooltipsAlert = mutations.pipe(
        first(),
        mapTo({
            type: 'nativeTooltips' as const,
            content: (
                <>
                    Sourcegraph has hidden {name || 'the code host'}'s native hover tooltips. You can toggle this at any
                    time: to enable the native tooltips run “Code host: prefer non-Sourcegraph hover tooltips” from the
                    command palette or set <code>"codeHost.useNativeTooltips": true</code> in your user settings.
                </>
            ),
        }),
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
                    element.classList.toggle(NATIVE_TOOLTIP_HIDDEN, !enabled)
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

export function registerNativeTooltipContributions(extensionsController: {
    services: Pick<Services, 'contribution'>
}): Unsubscribable {
    return extensionsController.services.contribution.registerContributions({
        contributions: {
            actions: [
                {
                    id: 'codeHost.toggleUseNativeTooltips',
                    command: 'updateConfiguration',
                    category: parseTemplate('Code host'),
                    commandArguments: [
                        parseTemplate('codeHost.useNativeTooltips'),
                        /* eslint-disable-next-line no-template-curly-in-string */
                        parseTemplate('${!config.codeHost.useNativeTooltips}'),
                        null,
                        parseTemplate('json'),
                    ],
                    title: parseTemplate(
                        /* eslint-disable-next-line no-template-curly-in-string */
                        'Prefer ${config.codeHost.useNativeTooltips && "Sourcegraph" || "non-Sourcegraph"} hover tooltips'
                    ),
                },
            ],
            menus: {
                commandPalette: [
                    {
                        action: 'codeHost.toggleUseNativeTooltips',
                    },
                ],
            },
        },
    })
}
