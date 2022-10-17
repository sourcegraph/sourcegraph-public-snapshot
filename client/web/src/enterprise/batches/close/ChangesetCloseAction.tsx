import React from 'react'

import { mdiCloseCircleOutline, mdiPlayCircleOutline } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

import styles from './ChangesetCloseAction.module.scss'

const iconClassNames = 'm-0 text-nowrap d-block d-sm-flex flex-column align-items-center justify-content-center'

export const ChangesetCloseActionClose: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames, styles.changesetCloseActionCloseFlash)}>
        <Tooltip content="This changeset will be closed on the code host when the batch change is closed.">
            <Icon
                aria-label="This changeset will be closed on the code host when the batch change is closed."
                svgPath={mdiCloseCircleOutline}
                inline={false}
            />
        </Tooltip>
        <span className="text-muted">Will close</span>
    </div>
)
export const ChangesetCloseActionKept: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div className={classNames(className, iconClassNames)}>
        <Tooltip content="This changeset will NOT be closed on the code host when the batch change is closed.">
            <Icon
                aria-label="This changeset will NOT be closed on the code host when the batch change is closed."
                svgPath={mdiPlayCircleOutline}
                inline={false}
            />
        </Tooltip>
        <span className="text-muted">Kept open</span>
    </div>
)
