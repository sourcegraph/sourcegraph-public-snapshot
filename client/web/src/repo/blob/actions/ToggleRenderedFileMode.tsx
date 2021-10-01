import EyeIcon from 'mdi-react/EyeIcon'
import React, { useEffect } from 'react'
import { useLocation } from 'react-router'

import { Tooltip } from '@sourcegraph/branded/src/components/tooltip/Tooltip'
import { ButtonLink } from '@sourcegraph/shared/src/components/LinkOrButton'
import { RenderMode } from '@sourcegraph/shared/src/util/url'

import { RepoHeaderContext } from '../../RepoHeader'

import { getURLForMode } from './utils'

interface ToggledRenderedFileModeProps {
    actionType: RepoHeaderContext['actionType']
    mode: RenderMode
}

/**
 * A repository header action that toggles between showing a rendered file and the file's original
 * source, for files that can be rendered (such as Markdown files).
 */
export const ToggleRenderedFileMode: React.FunctionComponent<ToggledRenderedFileModeProps> = ({ mode, actionType }) => {
    /**
     * The opposite mode of the current mode.
     * Used to enable switching between modes.
     */
    const otherMode: RenderMode = mode === 'code' ? 'rendered' : 'code'
    const label = mode === 'rendered' ? 'Show raw code file' : 'Show formatted file'
    const location = useLocation()

    useEffect(() => {
        Tooltip.forceUpdate()
    }, [mode])

    if (actionType === 'dropdown') {
        return (
            <ButtonLink className="btn repo-header__file-action" to={getURLForMode(location, otherMode)}>
                <EyeIcon className="icon-inline" />
                <span>{label}</span>
            </ButtonLink>
        )
    }

    return (
        <ButtonLink
            to={getURLForMode(location, otherMode)}
            data-tooltip={label}
            className="btn btn-icon repo-header__action"
        >
            <EyeIcon className="icon-inline" />{' '}
            <span className="d-none d-lg-inline ml-1">{mode === 'rendered' ? 'Raw' : 'Formatted'}</span>
        </ButtonLink>
    )
}
