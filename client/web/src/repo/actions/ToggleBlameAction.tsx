import { useCallback } from 'react'

import { mdiGit } from '@mdi/js'
import classNames from 'classnames'

import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { useBlameVisibility } from '../blame/useBlameVisibility'
import { RepoHeaderActionAnchor, RepoHeaderActionMenuLink } from '../components/RepoHeaderActions'

import styles from './ToggleBlameAction.module.scss'

interface Props {
    source?: 'repoHeader' | 'actionItemsBar'
    actionType?: 'nav' | 'dropdown'
}
export const ToggleBlameAction: React.FC<Props> = props => {
    const [isBlameVisible, setIsBlameVisible] = useBlameVisibility()

    const descriptiveText = `${isBlameVisible ? 'Hide' : 'Show'} Git blame line annotations`

    const toggleBlameState = useCallback(() => {
        if (isBlameVisible) {
            setIsBlameVisible(false)
            eventLogger.log('GitBlameDisabled')
        } else {
            setIsBlameVisible(true)
            eventLogger.log('GitBlameEnabled')
        }
    }, [isBlameVisible, setIsBlameVisible])

    const icon = (
        <Icon aria-hidden={true} svgPath={mdiGit} className={classNames(isBlameVisible && styles.iconActive)} />
    )

    if (props.source === 'actionItemsBar') {
        return (
            <SimpleActionItem tooltip={descriptiveText} isActive={isBlameVisible} onSelect={toggleBlameState}>
                {icon}
            </SimpleActionItem>
        )
    }

    if (props.actionType === 'dropdown') {
        return (
            <RepoHeaderActionMenuLink file={true} as={Button} onClick={toggleBlameState}>
                {icon}
                <span>{descriptiveText}</span>
            </RepoHeaderActionMenuLink>
        )
    }

    return (
        <Tooltip content={descriptiveText}>
            <RepoHeaderActionAnchor onSelect={toggleBlameState}>{icon}</RepoHeaderActionAnchor>
        </Tooltip>
    )
}
