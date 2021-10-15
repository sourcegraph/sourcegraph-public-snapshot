import classNames from 'classnames'
import React, { useCallback, useEffect, useMemo, useRef } from 'react'
import { Observable } from 'rxjs'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { useInputValidation, deriveInputClassName } from '@sourcegraph/shared/src/util/useInputValidation'

import { LINK_PROPS, URL_AUTH_ERROR, URL_FETCH_ERROR } from '../constants'

import { InfoText } from './InfoText'

const CheckIcon: React.FC<{ className?: string }> = ({ className }) => (
    <svg xmlns="http://www.w3.org/2000/svg" width="16" className={className} height="16" viewBox="0 0 8 8">
        <path
            fill="#37b24d"
            d="M2.3 6.73L.6 4.53c-.4-1.04.46-1.4 1.1-.8l1.1 1.4 3.4-3.8c.6-.63 1.6-.27 1.2.7l-4 4.6c-.43.5-.8.4-1.1.1z"
        />
    </svg>
)

const IconContainer: React.FC = ({ children }) => (
    <div className="options-page__icon-container position-absolute d-flex justify-content-center align-items-center">
        {children}
    </div>
)

export interface SourcegraphURLInputProps {
    label: string
    description: JSX.Element | string
    initialValue: string
    className?: string
    validate: (url: string) => Observable<string | undefined>
    editable?: boolean
    onChange?: (value: string) => void
    id: string
    dataTestId?: string
}

export const SourcegraphURLInput: React.FC<SourcegraphURLInputProps> = ({
    label,
    description,
    className,
    editable = true,
    initialValue,
    onChange,
    validate,
    id,
    dataTestId,
}) => {
    const urlInputReference = useRef<HTMLInputElement | null>(null)
    const [urlState, nextUrlFieldChange, nextUrlInputElement] = useInputValidation(
        useMemo(
            () => ({
                initialValue,
                synchronousValidators: [],
                asynchronousValidators: [validate],
            }),
            [initialValue, validate]
        )
    )
    const urlInputElements = useCallback(
        (urlInputElement: HTMLInputElement | null) => {
            urlInputReference.current = urlInputElement
            nextUrlInputElement(urlInputElement)
        },
        [nextUrlInputElement]
    )
    useEffect(() => {
        if (urlState.kind === 'VALID') {
            onChange?.(urlState.value)
        }
    }, [onChange, urlState])

    return (
        <div className={classNames('position-relative', className)}>
            <label htmlFor={id}>{label}</label>
            <div>
                <LoaderInput
                    loading={urlState.kind === 'LOADING' && !!urlState.value}
                    className={classNames(deriveInputClassName(urlState), {
                        'options-page__input-disabled': !editable,
                    })}
                >
                    <input
                        id={id}
                        type="url"
                        data-testid={dataTestId}
                        pattern="^https://.*"
                        placeholder="https://sourcegraph.example.com"
                        value={urlState.value}
                        onChange={nextUrlFieldChange}
                        ref={urlInputElements}
                        spellCheck={false}
                        disabled={!editable}
                        className={classNames(
                            'form-control',
                            'mb-2',
                            urlState.value ? deriveInputClassName(urlState) : ''
                        )}
                    />
                </LoaderInput>
                {!editable && <InfoText>{description}</InfoText>}
                {urlState.value ? (
                    <>
                        {urlState.kind === 'LOADING' && (
                            <>
                                {editable ? (
                                    <small className="text-muted d-block mt-1">Checking...</small>
                                ) : (
                                    <IconContainer>
                                        <LoadingSpinner className="options-page__icon-loading" />
                                    </IconContainer>
                                )}
                            </>
                        )}
                        {urlState.kind === 'VALID' && (
                            <>
                                {editable ? (
                                    <small className="valid-feedback" data-testid="test-valid-sourcegraph-url-feedback">
                                        Looks good!
                                    </small>
                                ) : (
                                    <IconContainer>
                                        <CheckIcon className="options-page__icon-check" />
                                    </IconContainer>
                                )}
                            </>
                        )}
                        {urlState.kind === 'INVALID' && (
                            <small className="invalid-feedback">
                                {urlState.reason === URL_FETCH_ERROR ? (
                                    'Failed to connect. Sourcegraph instance is down or incorrect URL address'
                                ) : urlState.reason === URL_AUTH_ERROR ? (
                                    <>
                                        Authentication to Sourcegraph failed.{' '}
                                        <a href={urlState.value} {...LINK_PROPS}>
                                            Sign in to your instance
                                        </a>{' '}
                                        to continue
                                    </>
                                ) : urlInputReference.current?.validity.typeMismatch ? (
                                    'Please enter a valid URL, including the protocol prefix (e.g. https://sourcegraph.example.com).'
                                ) : urlInputReference.current?.validity.patternMismatch ? (
                                    'The browser extension can only work over HTTPS in modern browsers.'
                                ) : (
                                    urlState.reason
                                )}
                            </small>
                        )}
                    </>
                ) : (
                    <InfoText>{description}</InfoText>
                )}
            </div>
        </div>
    )
}
