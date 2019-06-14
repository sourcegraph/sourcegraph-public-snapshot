import { isEqual } from 'lodash'
import { from, Observable, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, filter, map, publishReplay, refCount } from 'rxjs/operators'
import { parseTemplate } from '../../../../shared/src/api/client/context/expr/evaluator'
import { Services } from '../../../../shared/src/api/client/services'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { Settings } from '../../../../shared/src/settings/settings'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { isDefined, isNot } from '../../../../shared/src/util/types'
import { MutationRecordLike } from '../../shared/util/dom'
import { CodeHost } from './code_intelligence'
import { trackViews } from './views'

const NATIVE_TOOLTIP_HIDDEN = 'native-tooltip--hidden'

/**
 * Defines a native tooltip that is present on a page and exposes operations for manipulating it.
 */
export interface NativeTooltip {
    /** The native tooltip HTML element. */
    element: HTMLElement
}

/**
 * Handles added and removed native tooltips according to the {@link CodeHost} configuration.
 */
export function handleNativeTooltips(
    mutations: Observable<MutationRecordLike[]>,
    nativeTooltipsEnabled: Observable<boolean>,
    nativeTooltipResolvers: NonNullable<CodeHost['nativeTooltipResolvers']>
): Unsubscribable {
    /** A stream of added or removed native tooltips. */
    const nativeTooltips = mutations.pipe(trackViews(nativeTooltipResolvers))

    return nativeTooltips.subscribe(({ element, subscriptions }) => {
        subscriptions.add(
            nativeTooltipsEnabled
                // This subscription is correctly handled through the view's subscriptions.
                .subscribe(enabled => {
                    element.classList.toggle(NATIVE_TOOLTIP_HIDDEN, !enabled)
                })
        )
    })
}

export function nativeTooltipsEnabledFromSettings(settings: PlatformContext['settings']): Observable<boolean> {
    return from(settings).pipe(
        map(({ final }) => final),
        filter(isDefined),
        filter(isNot<ErrorLike | Settings, ErrorLike>(isErrorLike)),
        map(s => !!s['codeHost.useNativeTooltips']),
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
                        // tslint:disable-next-line no-invalid-template-strings
                        parseTemplate('${!config.codeHost.useNativeTooltips}'),
                        null,
                        parseTemplate('json'),
                    ],
                    title: parseTemplate(
                        // tslint:disable-next-line no-invalid-template-strings
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

export const nativeTooltipsAlert = (codeHostName?: string) =>
    `Sourcegraph has hidden ${codeHostName ||
        'the code host'}'s native hover tooltips. You can toggle this at any time: to enable the native tooltips run “Code host: prefer non-Sourcegraph hover tooltips” from the command palette or set \`"codeHost.useNativeTooltips": true\` in your user settings.`
