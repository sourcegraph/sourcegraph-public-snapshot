import { useCallback, useEffect } from 'react'

import { mdiGit } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'

import { SimpleActionItem } from '@sourcegraph/shared/src/actions/SimpleActionItem'
import { Icon } from '@sourcegraph/wildcard'

import { useExperimentalFeatures } from '../../stores'
import { eventLogger } from '../../tracking/eventLogger'
import { useBlameVisibility } from './useBlameVisibility'

import styles from './ToggleBlameAction.module.scss'

export const ToggleBlameAction: React.FC<{ location: H.Location }> = ({ location }) => {
    const extensionsAsCoreFeatures = useExperimentalFeatures(features => features.extensionsAsCoreFeatures)
    const [isBlameVisible, setIsBlameVisible] = useBlameVisibility()

    // Turn off visibility when the file path changes.
    useEffect(() => {
        setIsBlameVisible(false)
    }, [location.pathname, setIsBlameVisible])

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

    if (!extensionsAsCoreFeatures) {
        return null
    }

    return (
        <SimpleActionItem
            tooltip={descriptiveText}
            icon={
                <Icon aria-hidden={true} svgPath={mdiGit} className={classNames(isBlameVisible && styles.iconActive)} />
            }
            onClick={toggleBlameState}
        />
    )
}
