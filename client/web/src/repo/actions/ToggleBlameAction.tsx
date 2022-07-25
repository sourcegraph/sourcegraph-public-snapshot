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
            {/**
             * This <RepoHeaderActionButtonLink> must be wrapped with an additional span, since the tooltip currently has an issue that will
             * break its underlying <ButtonLink>'s onClick handler and it will no longer prevent the default page reload (with no href).
             */}
            <span>
                <RepoHeaderActionButtonLink
                    aria-label={descriptiveText}
                    onSelect={toggleBlameState}
                    className="btn-icon"
                >
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
