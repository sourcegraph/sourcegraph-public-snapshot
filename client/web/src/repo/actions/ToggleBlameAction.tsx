import { useCallback } from 'react'

import { mdiAccountDetails, mdiAccountDetailsOutline } from '@mdi/js'

import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { RenderMode } from '@sourcegraph/shared/src/util/url'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { useBlameVisibility } from '../blame/useBlameVisibility'
import { RepoHeaderActionAnchor, RepoHeaderActionMenuLink } from '../components/RepoHeaderActions'

interface Props {
    source?: 'repoHeader' | 'actionItemsBar'
    actionType?: 'nav' | 'dropdown'
    renderMode?: RenderMode
    isPackage: boolean
}
export const ToggleBlameAction: React.FC<Props & TelemetryV2Props> = props => {
    const [isBlameVisible, setIsBlameVisible] = useBlameVisibility(props.isPackage)

    const disabled = props.isPackage || props.renderMode === 'rendered'

    const descriptiveText = props.isPackage
        ? 'Git blame line annotations are not available when browsing packages'
        : disabled
        ? 'Git blame line annotations are not available when viewing a rendered document'
        : `${isBlameVisible ? 'Hide' : 'Show'} Git blame line annotations`

    const toggleBlameState = useCallback(() => {
        if (isBlameVisible) {
            setIsBlameVisible(false)
            props.telemetryRecorder.recordEvent('gitBlame', 'disabled')
            eventLogger.log('GitBlameDisabled')
        } else {
            setIsBlameVisible(true)
            props.telemetryRecorder.recordEvent('gitBlame', 'enabled')
            eventLogger.log('GitBlameEnabled')
        }
    }, [isBlameVisible, setIsBlameVisible, props.telemetryRecorder])

    const icon = (
        <Icon aria-hidden={true} svgPath={isBlameVisible && !disabled ? mdiAccountDetails : mdiAccountDetailsOutline} />
    )

    if (props.source === 'actionItemsBar') {
        return (
            <SimpleActionItem
                tooltip={descriptiveText}
                isActive={isBlameVisible}
                onSelect={toggleBlameState}
                disabled={disabled}
            >
                {icon}
            </SimpleActionItem>
        )
    }

    if (props.actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuLink file={true} as={Button} onClick={toggleBlameState} disabled={disabled}>
                {icon}
                <span>{descriptiveText}</span>
            </RepoHeaderActionMenuLink>
        )
    }

    return (
        <Tooltip content={descriptiveText}>
            <RepoHeaderActionAnchor onSelect={toggleBlameState} disabled={disabled}>
                {icon}
            </RepoHeaderActionAnchor>
        </Tooltip>
    )
}
