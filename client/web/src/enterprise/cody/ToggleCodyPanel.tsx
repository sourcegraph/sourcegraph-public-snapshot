import * as React from 'react'

import { Location, NavigateFunction } from 'react-router-dom'

import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import { Icon, Tooltip } from '@sourcegraph/wildcard'

import { BlobPanelTabID } from '../../repo/blob/panel/BlobPanel'
import { RepoHeaderActionButtonLink, RepoHeaderActionMenuLink } from '../../repo/components/RepoHeaderActions'
import { RepoHeaderContext } from '../../repo/RepoHeader'

import { codyIconPath } from './CodyIcon'

/**
 * A repository header action that toggles the visibility of the Cody panel.
 */
export const ToggleCodyPanel: React.FunctionComponent<
    {
        location: Location
        navigate: NavigateFunction
    } & RepoHeaderContext
> = ({ location, navigate, actionType }) => {
    const visible = parseQueryAndHash<BlobPanelTabID>(location.search, location.hash).viewState === 'cody'

    const title = `${visible ? 'Hide' : 'Show'} Cody`

    if (actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuLink file={true} href="#tab=cody">
                <Icon aria-hidden={true} svgPath={codyIconPath} />
                <span>{title}</span>
            </RepoHeaderActionMenuLink>
        )
    }
    return (
        <Tooltip content={title}>
            <RepoHeaderActionButtonLink
                aria-label={title}
                aria-controls="references-panel"
                aria-expanded={visible}
                file={false}
                to="#tab=cody"
            >
                <Icon aria-hidden={true} svgPath={codyIconPath} />
            </RepoHeaderActionButtonLink>
        </Tooltip>
    )
}
