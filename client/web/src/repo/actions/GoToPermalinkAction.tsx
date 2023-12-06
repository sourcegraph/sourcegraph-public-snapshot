import React, { useEffect, useMemo } from 'react'

import { mdiLink } from '@mdi/js'
import { useLocation, useNavigate } from 'react-router-dom'
import { fromEvent } from 'rxjs'
import { filter } from 'rxjs/operators'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isInputElement } from '@sourcegraph/shared/src/util/dom'
import { Icon, Link, Tooltip } from '@sourcegraph/wildcard'

import { replaceRevisionInURL } from '../../util/url'
import { RepoHeaderActionButtonLink, RepoHeaderActionMenuLink } from '../components/RepoHeaderActions'
import type { RepoHeaderContext } from '../RepoHeader'

interface GoToPermalinkActionProps extends RepoHeaderContext, TelemetryProps, TelemetryV2Props {
    /**
     * The current (possibly undefined or non-full-SHA) Git revision.
     */
    revision?: string

    /**
     * The commit SHA for the revision in the current location (URL).
     */
    commitID: string
}

/**
 * A repository header action that replaces the revision in the URL with the canonical 40-character
 * Git commit SHA.
 */
export const GoToPermalinkAction: React.FunctionComponent<GoToPermalinkActionProps> = props => {
    const { revision, commitID, actionType, repoName, telemetryService, telemetryRecorder } = props

    const navigate = useNavigate()
    const location = useLocation()
    const fullURL = location.pathname + location.search + location.hash
    const permalinkURL = useMemo(() => replaceRevisionInURL(fullURL, commitID), [fullURL, commitID])

    useEffect(() => {
        // Trigger the user presses 'y'.
        const subscription = fromEvent<KeyboardEvent>(window, 'keydown')
            .pipe(
                filter(
                    event =>
                        // 'y' shortcut (if no input element is focused)
                        event.key === 'y' && !!document.activeElement && !isInputElement(document.activeElement)
                )
            )
            .subscribe(event => {
                event.preventDefault()

                // Replace the revision in the current URL with the new one and push to history.
                navigate(permalinkURL)
            })

        return () => subscription.unsubscribe()
    }, [navigate, permalinkURL])

    if (revision === commitID) {
        return null // already at the permalink destination
    }

    const onClick = (): void => {
        telemetryService.log('PermalinkClicked', { repoName, commitID })
        telemetryRecorder.recordEvent('Permalink', 'clicked', {
            privateMetadata: {
                repoName,
                commitID,
            },
        })
    }

    if (actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuLink as={Link} file={true} to={permalinkURL} onSelect={onClick}>
                <Icon aria-hidden={true} svgPath={mdiLink} />
                <span>Permalink (with full Git commit SHA)</span>
            </RepoHeaderActionMenuLink>
        )
    }

    return (
        <Tooltip content="Permalink (with full Git commit SHA)">
            <RepoHeaderActionButtonLink aria-label="Permalink" file={false} to={permalinkURL} onSelect={onClick}>
                <Icon aria-hidden={true} svgPath={mdiLink} />
            </RepoHeaderActionButtonLink>
        </Tooltip>
    )
}
