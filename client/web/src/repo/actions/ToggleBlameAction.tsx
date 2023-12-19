import { useCallback } from 'react'

import { mdiGit } from '@mdi/js'

import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
import type { RenderMode } from '@sourcegraph/shared/src/util/url'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { useBlameVisibility } from '../blame/useBlameVisibility'
import { RepoHeaderActionAnchor, RepoHeaderActionMenuLink } from '../components/RepoHeaderActions'
import { RepoActionInfo } from '../RepoActionInfo'

interface Props {
    source?: 'repoHeader' | 'actionItemsBar'
    actionType?: 'nav' | 'dropdown'
    renderMode?: RenderMode
    isPackage: boolean
}

export const ToggleBlameAction: React.FC<Props> = props => {
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
            eventLogger.log('GitBlameDisabled')
        } else {
            setIsBlameVisible(true)
            eventLogger.log('GitBlameEnabled')
        }
    }, [isBlameVisible, setIsBlameVisible])

    const icon = <Icon aria-hidden={true} svgPath={mdiGit} />

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
            <RepoHeaderActionAnchor
                onSelect={toggleBlameState}
                disabled={disabled}
                className="d-flex justify-content-center align-items-center"
            >
                <RepoActionInfo displayName="Blame" icon={mdiGit} />
            </RepoHeaderActionAnchor>
        </Tooltip>
    )
}
