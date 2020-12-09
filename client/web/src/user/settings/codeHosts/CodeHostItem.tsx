import * as H from 'history'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CircleOutlineIcon from 'mdi-react/CircleOutlineIcon'
import React from 'react'
import { ExternalServiceKind } from '../../../graphql-operations'

interface CodeHostItemProps {
    /**
     * Title to show in the external service "button"
     */
    title: string

    /**
     * Icon to show in the external service "button"
     */
    icon: React.ComponentType<{ className?: string }>

    /**
     * A short description that will appear in the external service "button" under the title
     */
    shortDescription?: string

    kind: ExternalServiceKind

    className?: string
}

export const CodeHostItem: React.FunctionComponent<CodeHostItemProps> = ({
    title,
    icon: Icon,
    shortDescription,
    kind,
    className = '',
}) => (
    <div className={`p-2 d-flex align-items-start ${className}`}>
        <div className="align-self-center">
            <CircleOutlineIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon_off" />
            <AlertCircleIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon_warn" />
            <CheckCircleIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon_check" />
            <Icon className="icon-inline mb-0 mr-2" />
        </div>
        <div className="flex-1">
            <h3 className={shortDescription ? 'mb-0' : 'mt-1 mb-0'}>{title}</h3>
            {shortDescription && <p className="mb-0 text-muted">{shortDescription}</p>}
        </div>
        <div className="align-self-center">
            {true && (
                <button type="button" className="btn btn-success">
                    Connect
                </button>
            )}
            {true && (
                <button
                    type="button"
                    className="btn btn-link text-primary p-0 mr-2"
                    onClick={() => {}}
                    disabled={false}
                >
                    Edit
                </button>
            )}
            {true && (
                <button type="button" className="btn btn-link text-danger p-0" onClick={() => {}} disabled={false}>
                    Remove
                </button>
            )}
        </div>
    </div>
)
