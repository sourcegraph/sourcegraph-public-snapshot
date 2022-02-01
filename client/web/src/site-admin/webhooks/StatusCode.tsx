import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'

export interface Props {
    code: number
}

export const StatusCode: React.FunctionComponent<Props> = ({ code }) => (
    <span>
        <span className={classNames('mr-1')}>
            {code < 400 ? (
                <CheckIcon className="text-success icon-inline" />
            ) : (
                <AlertCircleIcon className="text-danger icon-inline" />
            )}
        </span>
        {code}
    </span>
)
