import { SourcegraphLogo } from '@sourcegraph/branded/src/components/SourcegraphLogo'
import { Container } from '@sourcegraph/wildcard'
import classNames from 'classnames'
import React from 'react'
import styles from './TrialStartFlowContainer.module.scss'

export const TrialStartFlowContainer: React.FunctionComponent<{
    afterOutside?: React.ReactNode
    children: React.ReactNode
}> = ({ afterOutside, children }) => (
    <div className={styles.page}>
        <Container className={styles.container}>
            <SourcegraphLogo className={classNames(styles.logo, 'mt-3 mb-4')} />
            {children}
        </Container>
        {afterOutside && <div className="mt-3">{afterOutside}</div>}
    </div>
)
