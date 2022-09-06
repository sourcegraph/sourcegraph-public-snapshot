import { useCallback, useEffect } from 'react'

import { mdiGit } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

import { useExperimentalFeatures } from '../../stores'
import { useBlameVisibility } from '../blame/useBlameVisibility'
import { RepoHeaderActionButtonLink } from '../components/RepoHeaderActions'

import styles from './ToggleBlameAction.module.scss'

export const ToggleBlameAction: React.FC<{
    filePath: string
}> = ({ filePath }) => {
    const extensionsAsCoreFeatures = useExperimentalFeatures(features => features.extensionsAsCoreFeatures)
    const [isBlameVisible, setIsBlameVisible] = useBlameVisibility()

    // Turn off visibility when the file path changes.
    useEffect(() => {
        setIsBlameVisible(false)
    }, [filePath, setIsBlameVisible])

    const descriptiveText = `${isBlameVisible ? 'Hide' : 'Show'} Git blame line annotations`

    const toggleBlameState = useCallback(() => setIsBlameVisible(!isBlameVisible), [isBlameVisible, setIsBlameVisible])

    if (!extensionsAsCoreFeatures) {
        return null
    }

    return (
        <Tooltip content={descriptiveText}>
            {/**
             * This <RepoHeaderActionButtonLink> must be wrapped with an additional span, since the tooltip currently has an issue that will
             * break its underlying <ButtonLink>'s onClick handler and it will no longer prevent the default page reload (with no href).
             */}
            <span>
                <RepoHeaderActionButtonLink aria-label={descriptiveText} onSelect={toggleBlameState}>
                    <Icon
                        aria-hidden={true}
                        svgPath={mdiGit}
                        className={classNames(isBlameVisible && styles.iconActive)}
                    />
                </RepoHeaderActionButtonLink>
            </span>
        </Tooltip>
    )
}
