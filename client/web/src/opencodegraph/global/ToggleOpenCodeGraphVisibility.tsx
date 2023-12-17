import { useCallback } from 'react'

import { mdiWeb, mdiWebOff } from '@mdi/js'

import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
import type { RenderMode } from '@sourcegraph/shared/src/util/url'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { RepoHeaderActionAnchor, RepoHeaderActionMenuLink } from '../../repo/components/RepoHeaderActions'

import { useOpenCodeGraphVisibility } from './useOpenCodeGraphVisibility'

interface Props {
    source?: 'repoHeader' | 'actionItemsBar'
    actionType?: 'nav' | 'dropdown'
    renderMode?: RenderMode
}

export const ToggleOpenCodeGraphVisibilityAction: React.FC<Props> = props => {
    const [visible, setVisible] = useOpenCodeGraphVisibility()

    const disabled = props.renderMode === 'rendered'
    const descriptiveText = disabled
        ? 'OpenCodeGraph metadata is not available in rendered files'
        : `${visible ? 'Hide' : 'Show'} OpenCodeGraph metadata`

    const onCycle = useCallback(() => {
        setVisible(prevVisible => !prevVisible)
        // TODO(sqs): telemetry
    }, [setVisible])

    const icon = disabled || !visible ? mdiWebOff : mdiWeb

    if (props.source === 'actionItemsBar') {
        return (
            <SimpleActionItem tooltip={descriptiveText} isActive={visible} onSelect={onCycle} disabled={disabled}>
                <Icon aria-hidden={true} svgPath={icon} />
            </SimpleActionItem>
        )
    }

    if (props.actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuLink file={true} as={Button} onClick={onCycle} disabled={disabled}>
                <Icon aria-hidden={true} svgPath={icon} />
                <span>{descriptiveText}</span>
            </RepoHeaderActionMenuLink>
        )
    }

    return (
        <Tooltip content={descriptiveText}>
            <RepoHeaderActionAnchor
                onSelect={onCycle}
                disabled={disabled}
                className="d-flex justify-content-center align-items-center"
            >
                <Icon aria-hidden={true} svgPath={icon} />
            </RepoHeaderActionAnchor>
        </Tooltip>
    )
}
