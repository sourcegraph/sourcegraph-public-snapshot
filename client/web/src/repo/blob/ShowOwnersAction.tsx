import { FunctionComponent, useEffect, useState } from 'react'

import { mdiAccountOutline } from '@mdi/js'
import { History, Location } from 'history'

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

const ShowOwnersAction: FunctionComponent<
    {
        location: Location
        history: History
    } & RepoHeaderContext
> = props => {
    const [visible, setVisible] = useState<boolean>(false)
    useEffect(() => {
        const parsedQuery = parseQueryAndHash<BlobPanelTabID>(props.location.search, props.location.hash)
        setVisible(parsedQuery.viewState === 'ownership')
    }, [props.location.search, props.location.hash])

    const toggle = (): void => {
        const parsedQuery = parseQueryAndHash<BlobPanelTabID>(props.location.search, props.location.hash)
        if (!visible) {
            parsedQuery.viewState = 'ownership' // defaults to last-viewed tab, or first tab
        } else {
            delete parsedQuery.viewState
        }
        const lineRangeQueryParameter = toPositionOrRangeQueryParameter({ range: lprToRange(parsedQuery) })
        props.history.push({
            search: formatSearchParameters(
                addLineRangeQueryParameter(new URLSearchParams(location.search), lineRangeQueryParameter)
            ),
            hash: toViewStateHash(parsedQuery.viewState),
        })
    }

    const descriptiveText = `${visible ? 'Hide' : 'Show'} ownership`

    if (props.actionType === 'dropdown') {
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
export default ShowOwnersAction
