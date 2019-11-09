import { upperFirst } from 'lodash'
import * as React from 'react'
import { ERAUTHREQUIRED } from '../../../../shared/src/backend/errors'
import { EINVALIDSOURCEGRAPHURL } from '../../shared/util/context'
import { ConnectionStatus } from './useSourcegraphURL'

const statusClassNames: Record<ConnectionStatus['type'], string> = {
    connecting: 'bg-warning',
    connected: 'bg-success',
    error: 'bg-danger',
}

/**
 * This is the [Word-Joiner](https://en.wikipedia.org/wiki/Word_joiner) character.
 * We are using this as a &nbsp; that has no width to maintain line height when the
 * url is being updated (therefore no text is in the status indicator).
 */
const zeroWidthNbsp = '\u2060'

export interface ServerURLFormProps {
    className?: string
    connectionStatus: undefined | ConnectionStatus
    sourcegraphURL: string
    onSourcegraphURLChange: (value: string) => void
    onSourcegraphURLSubmit: () => void
    requestSourcegraphURLPermissions: (e: React.MouseEvent) => void
}

export const ServerURLForm: React.FunctionComponent<ServerURLFormProps> = ({
    connectionStatus,
    onSourcegraphURLChange,
    onSourcegraphURLSubmit,
    requestSourcegraphURLPermissions,
    sourcegraphURL,
    className,
}) => {
    const onURLChange = React.useCallback(
        ({ target: { value } }: React.ChangeEvent<HTMLInputElement>): void => {
            onSourcegraphURLChange(value)
        },
        [onSourcegraphURLChange]
    )
    const onURLSubmit = React.useCallback(
        (event: React.FormEvent<HTMLFormElement>): void => {
            event.preventDefault()
            onSourcegraphURLSubmit()
        },
        [onSourcegraphURLSubmit]
    )
    const statusString = connectionStatus ? upperFirst(connectionStatus.type) : zeroWidthNbsp
    const statusClassName = connectionStatus ? statusClassNames[connectionStatus.type] : 'bg-secondary'
    return (
        // eslint-disable-next-line react/forbid-elements
        <form className={'server-url-form ' + (className || '')} onSubmit={onURLSubmit}>
            <label htmlFor="sourcegraph-url">Sourcegraph URL</label>
            <div className="input-group">
                <div className="input-group-prepend">
                    <span className="input-group-text">
                        <span>
                            <span className={'server-url-form__status-indicator ' + statusClassName} />{' '}
                            <span className="e2e-connection-status">{statusString}</span>
                        </span>
                    </span>
                </div>
                <input
                    type="text"
                    className="form-control e2e-sourcegraph-url"
                    id="sourcegraph-url"
                    value={sourcegraphURL}
                    onChange={onURLChange}
                    spellCheck={false}
                    autoCapitalize="off"
                    autoCorrect="off"
                />
            </div>
            {connectionStatus &&
                connectionStatus.type === 'error' &&
                (connectionStatus.error.code === EINVALIDSOURCEGRAPHURL ? (
                    <div className="mt-1">Invalid Sourcegraph URL.</div>
                ) : connectionStatus.error.code === ERAUTHREQUIRED ? (
                    <div className="mt-1">
                        Authentication to Sourcegraph failed.{' '}
                        <a href={sourcegraphURL} target="_blank" rel="noopener noreferrer">
                            Sign in to your instance
                        </a>{' '}
                        to continue.
                    </div>
                ) : (
                    <div className="mt-1">
                        <p>
                            Unable to connect to{' '}
                            <a href={sourcegraphURL} target="_blank" rel="noopener noreferrer">
                                {sourcegraphURL}
                            </a>
                            . Ensure the URL is correct and you are{' '}
                            <a href={sourcegraphURL + '/sign_in'} target="_blank" rel="noopener noreferrer">
                                logged in
                            </a>
                            .
                        </p>
                        {connectionStatus.urlHasPermissions === false && (
                            <p>
                                You may need to{' '}
                                <a href="#" onClick={requestSourcegraphURLPermissions}>
                                    grant the Sourcegraph browser extension additional permissions
                                </a>{' '}
                                for this URL.
                            </p>
                        )}
                        <p>
                            <b>Site admins:</b> ensure that{' '}
                            <a
                                href="https://docs.sourcegraph.com/admin/config/site_config"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                all users can create access tokens
                            </a>
                            .
                        </p>
                    </div>
                ))}
        </form>
    )
}
