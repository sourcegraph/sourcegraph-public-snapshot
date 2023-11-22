import React from 'react'

import { mdiEye } from '@mdi/js'
import { useLocation } from 'react-router-dom'

import type { RenderMode } from '@sourcegraph/shared/src/util/url'
import { createLinkUrl, Icon, Link, Tooltip } from '@sourcegraph/wildcard'

import { RepoHeaderActionButtonLink, RepoHeaderActionMenuLink } from '../../components/RepoHeaderActions'
import type { RepoHeaderContext } from '../../RepoHeader'

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

    if (actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuLink as={Link} to={createLinkUrl(getURLForMode(location, otherMode))} file={true}>
                <Icon aria-hidden={true} svgPath={mdiEye} />
                <span>{label}</span>
            </RepoHeaderActionMenuLink>
        )
    }

    return (
        <Tooltip content={label}>
            <RepoHeaderActionButtonLink file={false} to={createLinkUrl(getURLForMode(location, otherMode))}>
                <Icon aria-hidden={true} svgPath={mdiEye} />{' '}
                <span className="d-none d-lg-inline ml-1">{mode === 'rendered' ? 'Raw' : 'Formatted'}</span>
            </RepoHeaderActionButtonLink>
        </Tooltip>
    )
}
