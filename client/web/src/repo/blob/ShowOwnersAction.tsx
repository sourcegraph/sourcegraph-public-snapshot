import { FunctionComponent, useEffect, useState } from 'react'

import { mdiAccountOutline } from '@mdi/js'
import { useHistory, useLocation } from 'react-router'

import {
    addLineRangeQueryParameter,
    formatSearchParameters,
    lprToRange,
    toPositionOrRangeQueryParameter,
    toViewStateHash,
} from '@sourcegraph/common'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import { Icon, Tooltip } from '@sourcegraph/wildcard'

import { RepoHeaderActionButtonLink, RepoHeaderActionMenuItem } from '../components/RepoHeaderActions'
import { RepoHeaderContext } from '../RepoHeader'

import { BlobPanelTabID } from './panel/BlobPanel'

export const ShowOwnersAction: FunctionComponent<Pick<RepoHeaderContext, 'actionType'>> = ({ actionType }) => {
    const location = useLocation()
    const history = useHistory()

    const [visible, setVisible] = useState<boolean>(false)
    useEffect(() => {
        const parsedQuery = parseQueryAndHash<BlobPanelTabID>(location.search, location.hash)
        setVisible(parsedQuery.viewState === 'ownership')
    }, [location.search, location.hash])

    const toggle = (): void => {
        const parsedQuery = parseQueryAndHash<BlobPanelTabID>(location.search, location.hash)
        if (!visible) {
            parsedQuery.viewState = 'ownership' // defaults to last-viewed tab, or first tab
        } else {
            delete parsedQuery.viewState
        }
        const lineRangeQueryParameter = toPositionOrRangeQueryParameter({ range: lprToRange(parsedQuery) })
        history.push({
            search: formatSearchParameters(
                addLineRangeQueryParameter(new URLSearchParams(location.search), lineRangeQueryParameter)
            ),
            hash: toViewStateHash(parsedQuery.viewState),
        })
    }

    const descriptiveText = `${visible ? 'Hide' : 'Show'} ownership`

    if (actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuItem file={true} onSelect={toggle}>
                <Icon aria-hidden={true} svgPath={mdiAccountOutline} />
                <span>{descriptiveText}</span>
            </RepoHeaderActionMenuItem>
        )
    }
    return (
        <Tooltip content={descriptiveText}>
            <RepoHeaderActionButtonLink
                aria-label={descriptiveText}
                aria-controls="references-panel"
                aria-expanded={visible}
                onSelect={toggle}
                file={true}
            >
                <Icon aria-hidden={true} svgPath={mdiAccountOutline} />
            </RepoHeaderActionButtonLink>
        </Tooltip>
    )
}
