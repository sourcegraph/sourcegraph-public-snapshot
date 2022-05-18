import React from 'react'

import CommentAlert from 'mdi-react/CommentAlertIcon'

import { Link } from '@sourcegraph/wildcard'

import styles from './ReferencesPanelFeedbackCta.module.scss'

export const ReferencesPanelFeedbackCta: React.FunctionComponent = () => (
    <div className={styles.container}>
        <CommentAlert size={16} />
        <Link to="https://github.com/sourcegraph/sourcegraph/issues">Send us your reference panel feedback!</Link>
    </div>
)
