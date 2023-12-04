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
            {code < 400 && code > 0 ? (
                <Icon className="text-success" aria-label="Success" svgPath={mdiCheck} />
            ) : (
                <Icon className="text-danger" aria-label="Failed" svgPath={mdiAlertCircle} />
            )}
        </span>
        {code > 0 ? code : 'Network error'}
    </span>
)
