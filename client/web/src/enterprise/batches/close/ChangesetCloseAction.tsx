import React from 'react'

import classNames from 'classnames'
import CloseCircleOutlineIcon from 'mdi-react/CloseCircleOutlineIcon'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'

import styles from './ChangesetCloseAction.module.scss'

const iconClassNames = 'm-0 text-nowrap d-block d-sm-flex flex-column align-items-center justify-content-center'

export const ChangesetCloseActionClose: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames, styles.changesetCloseActionCloseFlash)}>
        <CloseCircleOutlineIcon data-tooltip="This changeset will be closed on the code host when the batch change is closed." />
        <span className="text-muted">Will close</span>
    </div>
)
export const ChangesetCloseActionKept: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <PlayCircleOutlineIcon data-tooltip="This changeset will NOT be closed on the code host when the batch change is closed." />
        <span className="text-muted">Kept open</span>
    </div>
)
