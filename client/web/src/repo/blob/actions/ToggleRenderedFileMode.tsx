import React, { useEffect } from 'react'

import EyeIcon from 'mdi-react/EyeIcon'
import { useLocation } from 'react-router'

import { RenderMode } from '@sourcegraph/shared/src/util/url'
import { TooltipController, Icon } from '@sourcegraph/wildcard'

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
export const ToggleRenderedFileMode: React.FunctionComponent<React.PropsWithChildren<ToggledRenderedFileModeProps>> = ({
    mode,
    actionType,
}) => {
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
            <RepoHeaderActionButtonLink to={getURLForMode(location, otherMode)} file={true}>
                <Icon role="img" as={EyeIcon} aria-hidden="true" />
                <span>{label}</span>
            </RepoHeaderActionButtonLink>
        )
    }

    return (
        <RepoHeaderActionButtonLink
            className="btn-icon"
            file={false}
            to={getURLForMode(location, otherMode)}
            data-tooltip={label}
        >
            <Icon role="img" as={EyeIcon} aria-hidden={true} />{' '}
            <span className="d-none d-lg-inline ml-1">{mode === 'rendered' ? 'Raw' : 'Formatted'}</span>
        </RepoHeaderActionButtonLink>
    )
}
