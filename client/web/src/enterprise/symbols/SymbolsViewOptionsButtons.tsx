import React, { useCallback } from 'react'
import H from 'history'
import { ButtonLink } from '../../../../shared/src/components/LinkOrButton'
import { useSymbolsViewOptions } from './useSymbolsViewOptions'
import { Toggle } from '../../../../branded/src/components/Toggle'
import { Link } from 'react-router-dom'

interface ButtonProps {
    toggleURL: H.LocationDescriptorObject
    value: boolean
    history: H.History
    className?: string
}

const SymbolsViewOptionsToggle: React.FunctionComponent<ButtonProps> = ({
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

interface Props extends ReturnType<typeof useSymbolsViewOptions> {
    history: H.History
}

export const SymbolsInternalsViewOptionToggle: React.FunctionComponent<Props> = ({
    viewOptions,
    toggleURLs,
    history,
}) => (
    <SymbolsViewOptionsToggle toggleURL={toggleURLs.internals} value={viewOptions.internals} history={history}>
        Internals
    </SymbolsViewOptionsToggle>
)

export const SymbolsExternalsViewOptionToggle: React.FunctionComponent<Props> = ({
    viewOptions,
    toggleURLs,
    history,
}) => (
    <SymbolsViewOptionsToggle toggleURL={toggleURLs.externals} value={viewOptions.externals} history={history}>
        External usage
    </SymbolsViewOptionsToggle>
)
