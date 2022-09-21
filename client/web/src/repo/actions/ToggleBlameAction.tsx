import { useCallback } from 'react'

import { mdiGit } from '@mdi/js'
import classNames from 'classnames'

import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
import { Icon } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { useBlameVisibility } from '../blame/useBlameVisibility'

import styles from './ToggleBlameAction.module.scss'

export const ToggleBlameAction: React.FC = () => {
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

    return (
        <SimpleActionItem isActive={isBlameVisible} tooltip={descriptiveText} onSelect={toggleBlameState}>
            <Icon
                aria-hidden={true}
                svgPath={mdiGit}
                className={classNames(styles.icon, isBlameVisible && styles.iconActive)}
            />
        </SimpleActionItem>
    )
}
