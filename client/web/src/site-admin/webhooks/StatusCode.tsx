import React from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'

import { Icon } from '@sourcegraph/wildcard'

export interface Props {
    code: number
}

export const StatusCode: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ code }) => (
    <span>
        <span className={classNames('mr-1')}>
            {code < 400 ? (
                <Icon className="text-success" as={CheckIcon} />
            ) : (
                <Icon className="text-danger" as={AlertCircleIcon} />
            )}
        </span>
        {code}
    </span>
)
