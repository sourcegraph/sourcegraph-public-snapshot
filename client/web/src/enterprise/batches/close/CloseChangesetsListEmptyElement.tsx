import classNames from 'classnames'
import React from 'react'

import styles from './CloseChangesetsListEmptyElement.module.scss'

export const CloseChangesetsListEmptyElement: React.FunctionComponent<{}> = () => (
    <div className="col-md-8 offset-md-2 col-sm-12 card mt-5">
        <div className={classNames(styles.closeChangesetsListEmptyElementBody, 'card-body p-5')}>
            <h2 className="text-center font-weight-normal">
                Closing this batch change will not alter changesets and no changesets will remain open.
            </h2>
        </div>
    </div>
)
