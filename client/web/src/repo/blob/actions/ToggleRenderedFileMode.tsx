import EyeIcon from 'mdi-react/EyeIcon'
import React, { useEffect } from 'react'
import { useLocation } from 'react-router'

import { RenderMode } from '@sourcegraph/shared/src/util/url'
import { TooltipController } from '@sourcegraph/wildcard'

import { RepoHeaderActionButtonLink } from '../../components/RepoHeaderActions'
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
        TooltipController.forceUpdate()
    }, [mode])

    if (actionType === 'dropdown') {
        return (
            <RepoHeaderActionButtonLink className="btn" to={getURLForMode(location, otherMode)} file={true}>
                <EyeIcon className="icon-inline" />
                <span>{label}</span>
            </RepoHeaderActionButtonLink>
        )
    }

    return (
        <RepoHeaderActionButtonLink
            className="btn btn-icon"
            file={false}
            to={getURLForMode(location, otherMode)}
            data-tooltip={label}
        >
            <EyeIcon className="icon-inline" />{' '}
            <span className="d-none d-lg-inline ml-1">{mode === 'rendered' ? 'Raw' : 'Formatted'}</span>
        </RepoHeaderActionButtonLink>
    )
}
