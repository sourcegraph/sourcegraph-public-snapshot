import React from 'react'

import classNames from 'classnames'
import CommentAlert from 'mdi-react/CommentAlertIcon'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Link } from '@sourcegraph/wildcard'

import styles from './ReferencesPanelFeedbackCta.module.scss'

export const ReferencesPanelFeedbackCta: React.FunctionComponent = () => {
    const [enabled] = useTemporarySetting('codeintel.referencePanel.redesign.enabled', false)

    return (
        <>
            {enabled === true && (
                <div className={classNames('m-0 p-0 pr-3', styles.container)}>
                    <CommentAlert size={16} />
                    <Link to="https://github.com/sourcegraph/sourcegraph/discussions/35668" className="ml-2">
                        Send us your reference panel feedback!
                    </Link>
                </div>
            )}
        </>
    )
}
