import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'

interface Props extends React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement> {
    loading: boolean
    label: string
}

export const LoaderButton: React.FunctionComponent<Props> = props => (
    // eslint-disable-next-line react/button-has-type
    <button type={props.type ?? 'button'} {...props}>
        {props.loading ? <LoadingSpinner className="icon-inline" /> : props.label}
    </button>
)
