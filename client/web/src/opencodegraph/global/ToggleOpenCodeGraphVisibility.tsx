import { useCallback } from 'react'

import { mdiWeb, mdiWebOff } from '@mdi/js'

import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { RenderMode } from '@sourcegraph/shared/src/util/url'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { RepoHeaderActionAnchor, RepoHeaderActionMenuLink } from '../../repo/components/RepoHeaderActions'

import { useOpenCodeGraphVisibility } from './useOpenCodeGraphVisibility'

import styles from './ToggleOpenCodeGraphVisibility.module.scss'

interface Props extends TelemetryV2Props {
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
        props.telemetryRecorder.recordEvent('blob.openCodeGraph.metadata', visible ? 'hide' : 'show')
        setVisible(prevVisible => !prevVisible)
    }, [setVisible, visible, props.telemetryRecorder])

    const icon = disabled || !visible ? mdiWebOff : mdiWeb

    if (props.source === 'actionItemsBar') {
        return (
            <SimpleActionItem tooltip={descriptiveText} isActive={visible} onSelect={onCycle} disabled={disabled}>
                <Icon aria-hidden={true} svgPath={icon} className={styles.repoActionIcon} />
            </SimpleActionItem>
        )
    }

    if (props.actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuLink file={true} as={Button} onClick={onCycle} disabled={disabled}>
                <Icon aria-hidden={true} svgPath={icon} className={styles.repoActionIcon} />
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
                <Icon aria-hidden={true} svgPath={icon} className={styles.repoActionIcon} />
            </RepoHeaderActionAnchor>
        </Tooltip>
    )
}
