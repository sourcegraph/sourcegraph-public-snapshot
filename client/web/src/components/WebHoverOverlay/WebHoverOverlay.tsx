import React, { useEffect } from 'react'

import classNames from 'classnames'
import { Location, useLocation } from 'react-router-dom'
import { fromEvent, Observable } from 'rxjs'
import { finalize, tap } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/common'
import { urlForClientCommandOpen } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverOverlay, HoverOverlayProps } from '@sourcegraph/shared/src/hover/HoverOverlay'
import { Settings, SettingsCascadeOrError, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { HoverThresholdProps } from '../../repo/RepoContainer'

import styles from './WebHoverOverlay.module.scss'

export interface WebHoverOverlayProps
    extends Omit<
            HoverOverlayProps,
            'className' | 'closeButtonClassName' | 'actionItemClassName' | 'actionItemStyleProps'
        >,
        HoverThresholdProps,
        SettingsCascadeProps {
    hoveredTokenElement?: HTMLElement
    /**
     * If the hovered token doesn't have a corresponding DOM element, this prop
     * can be used to trigger the "click to go to definition" functionality.
     */
    hoveredTokenClick?: Observable<unknown>
    nav?: (url: string) => void

    hoverOverlayContainerClassName?: string
}

export const WebHoverOverlay: React.FunctionComponent<React.PropsWithChildren<WebHoverOverlayProps>> = props => {
    const location = useLocation()
    const { hoverOrError, onHoverShown, hoveredToken } = props

    /** Whether the hover has actual content (that provides value to the user) */
    const hoverHasValue = hoverOrError !== 'loading' && !isErrorLike(hoverOrError) && !!hoverOrError?.contents?.length

    useEffect(() => {
        if (hoverHasValue) {
            onHoverShown?.()
        }
    }, [hoveredToken?.filePath, hoveredToken?.line, hoveredToken?.character, onHoverShown, hoverHasValue])

    const clickToGoToDefinition = getClickToGoToDefinition(props.settingsCascade)

    useEffect(() => {
        if (!clickToGoToDefinition) {
            return
        }

        const token = props.hoveredTokenElement
        const click = props.hoveredTokenClick ?? (token ? fromEvent(token, 'click') : null)
        const nav = props.nav
        if (!click || !nav) {
            return
        }

        const urlAndType = getGoToURL(props.actionsOrError, location)
        if (!urlAndType) {
            return
        }
        const { url, actionType } = urlAndType

        const oldCursor = token?.style.cursor
        if (token) {
            token.style.cursor = 'pointer'
        }

        const subscription = click
            .pipe(
                tap(() => {
                    const selection = window.getSelection()
                    if (selection !== null && selection.toString() !== '') {
                        return
                    }

                    props.telemetryService.log(`${actionType}HoverOverlay.click`)
                    nav(url)
                }),
                finalize(() => {
                    if (token && oldCursor) {
                        token.style.cursor = oldCursor
                    }
                })
            )
            .subscribe()

        return () => subscription.unsubscribe()
    }, [
        props.actionsOrError,
        props.hoveredTokenElement,
        props.hoveredTokenClick,
        location,
        props.nav,
        props.telemetryService,
        clickToGoToDefinition,
        hoveredToken,
    ])

    return (
        <HoverOverlay
            {...props}
            className={classNames(styles.webHoverOverlay, props.hoverOverlayContainerClassName)}
            closeButtonClassName={styles.webHoverOverlayCloseButton}
            actionItemClassName="border-0"
            actionItemStyleProps={{
                actionItemSize: 'sm',
                actionItemVariant: 'secondary',
            }}
        />
    )
}

WebHoverOverlay.displayName = 'WebHoverOverlay'

/**
 * Returns the URL and type to perform the "go to ..." navigation.
 */
export const getGoToURL = (
    actionsOrError: WebHoverOverlayProps['actionsOrError'],
    location: Location
): {
    url: string
    actionType: 'definition' | 'reference'
} | null => {
    const definitionAction =
        Array.isArray(actionsOrError) &&
        actionsOrError.find(a => a.action.id === 'goToDefinition.preloaded' && !a.disabledWhen)

    const referenceAction =
        Array.isArray(actionsOrError) && actionsOrError.find(a => a.action.id === 'findReferences' && !a.disabledWhen)

    const action = definitionAction || referenceAction
    if (!action) {
        return null
    }

    const url = urlForClientCommandOpen(action.action, location.hash)
    if (!url) {
        return null
    }

    return { url, actionType: action === definitionAction ? 'definition' : 'reference' }
}

export const getClickToGoToDefinition = (settingsCascade: SettingsCascadeOrError<Settings>): boolean => {
    if (settingsCascade.final && !isErrorLike(settingsCascade.final)) {
        const value = settingsCascade.final['codeIntelligence.clickToGoToDefinition'] as boolean
        return value ?? true
    }
    return true
}
