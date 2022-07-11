import React from 'react'

import { mdiCheck, mdiAlertCircle } from '@mdi/js'
import classNames from 'classnames'

import { Icon } from '@sourcegraph/wildcard'

export interface Props {
    code: number
}

export const StatusCode: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ code }) => (
    <span>
        <span className={classNames('mr-1')}>
            {code < 400 ? (
                <Icon className="text-success" aria-hidden={true} svgPath={mdiCheck} />
            ) : (
                <Icon className="text-danger" aria-hidden={true} svgPath={mdiAlertCircle} />
            )}
        </span>
        {code}
    </span>
)
