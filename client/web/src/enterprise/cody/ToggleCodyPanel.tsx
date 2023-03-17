import * as React from 'react'

import { Location } from 'react-router-dom'

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
    } & RepoHeaderContext
> = ({ location, actionType }) => {
    const visible = parseQueryAndHash<BlobPanelTabID>(location.search, location.hash).viewState === 'cody'
    const href = visible ? '#' : '#tab=cody'

    const title = `${visible ? 'Hide' : 'Show'} Cody`

    if (actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuLink file={true} href={href}>
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
                to={href}
            >
                <Icon aria-hidden={true} svgPath={codyIconPath} />
            </RepoHeaderActionButtonLink>
        </Tooltip>
    )
}
