import { useCallback } from 'react'

import { mdiGit } from '@mdi/js'
import classNames from 'classnames'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Icon, Tooltip } from '@sourcegraph/wildcard'

import { RepoHeaderActionButtonLink } from '../components/RepoHeaderActions'

import styles from './ToggleBlameAction.module.scss'

export const ToggleBlameAction: React.FC = () => {
    const [isBlameVisible, setValueAndSave] = useTemporarySetting('git.showBlame', false)

    const descriptiveText = isBlameVisible
        ? 'Hide Git blame line annotations'
        : 'Show Git blame line annotations for the whole file'

    const toggleBlameState = useCallback(() => setValueAndSave(isVisible => !isVisible), [setValueAndSave])

    return (
        <Tooltip content={descriptiveText}>
            <span>
                <RepoHeaderActionButtonLink
                    aria-label={descriptiveText}
                    onSelect={toggleBlameState}
                    className="btn-icon"
                >
                    <Icon
                        aria-hidden={true}
                        svgPath={mdiGit}
                        className={classNames(styles.icon, isBlameVisible && styles.iconActive)}
                    />
                </RepoHeaderActionButtonLink>
            </span>
            {/* <Button className={styles.item} onClick={toggleBlameState}>
                <img src={iconURL} alt="Git blame" className={styles.icon} />
            </Button> */}
        </Tooltip>
    )
}
