import H from 'history'
import React, { useCallback } from 'react'

import { Toggle } from '../../../../branded/src/components/Toggle'
import { ButtonLink } from '../../../../shared/src/components/LinkOrButton'

import { useContextViewOptions } from './useContextViewOptions'

interface ButtonProps {
    toggleURL: H.LocationDescriptorObject
    value: boolean
    history: H.History
    className?: string
}

const ContextViewOptionsToggle: React.FunctionComponent<ButtonProps> = ({
    toggleURL,
    value,
    history,
    className = '',
    children,
}) => {
    const onToggle = useCallback(() => {
        history.replace(toggleURL)
    }, [history, toggleURL])
    return (
        <>
            <Toggle value={value} onToggle={onToggle} className="d-flex align-self-center ml-2" />
            <ButtonLink
                to={toggleURL}
                pressed={value}
                className={`btn-sm align-self-center mt-1 text-decoration-none ${className}`}
            >
                {children}
            </ButtonLink>
        </>
    )
}

interface Props extends ReturnType<typeof useContextViewOptions> {
    history: H.History
}

export const ContextInternalsViewOptionToggle: React.FunctionComponent<Props> = ({
    viewOptions,
    toggleURLs,
    history,
}) => (
    <ContextViewOptionsToggle toggleURL={toggleURLs.internals} value={viewOptions.internals} history={history}>
        Internals
    </ContextViewOptionsToggle>
)

export const ContextExternalsViewOptionToggle: React.FunctionComponent<Props> = ({
    viewOptions,
    toggleURLs,
    history,
}) => (
    <ContextViewOptionsToggle toggleURL={toggleURLs.externals} value={viewOptions.externals} history={history}>
        External usage
    </ContextViewOptionsToggle>
)
