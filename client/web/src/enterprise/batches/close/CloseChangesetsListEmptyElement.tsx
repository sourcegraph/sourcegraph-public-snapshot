import React from 'react'

import { Text } from '@sourcegraph/wildcard'

import styles from './CloseChangesetsListEmptyElement.module.scss'

export const CloseChangesetsListEmptyElement: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <div className={styles.closeChangesetsListEmptyElementBody}>
        <Text alignment="center" weight="regular" className="text-muted">
            Closing this batch change will not alter changesets and no changesets will remain open.
        </Text>
    </div>
)
