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
                <Icon role="img" className="text-success" as={CheckIcon} aria-hidden={true} />
            ) : (
                <Icon role="img" className="text-danger" as={AlertCircleIcon} aria-hidden={true} />
            )}
        </span>
        {code}
    </span>
)
