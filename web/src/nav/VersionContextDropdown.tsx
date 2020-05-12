import React, { useMemo } from 'react'
import { ListboxOption, ListboxInput, ListboxButton, ListboxPopover, ListboxList } from '@reach/listbox'
import { useObservable } from '../../../shared/src/util/useObservable'
import { fetchVersionContexts } from '../search/backend'
import { catchError } from 'rxjs/operators'
import { of } from 'rxjs'
import { ErrorLike, asError, isErrorLike } from '../../../shared/src/util/errors'
import * as GQL from '../../../shared/src/graphql/schema'
import classNames from 'classnames'
import { VersionContextProps } from '../../../shared/src/search/util'

export const VersionContextDropdown: React.FunctionComponent<VersionContextProps> = (props: VersionContextProps) => {
    const versionContexts: GQL.IVersionContext[] | ErrorLike | undefined = useObservable(
        useMemo(() => fetchVersionContexts().pipe(catchError(err => of<ErrorLike>(asError(err)))), [])
    )

    const updateValue = (newValue: string): void => {
        props.setVersionContext(newValue)
    }

    // const sortVersionContexts = (versionContexts: [])
    if (!versionContexts || (!isErrorLike(versionContexts) && versionContexts.length === 0)) {
        return null
    }

    return (
        <>
            {versionContexts && !isErrorLike(versionContexts) && versionContexts.length > 0 ? (
                <div className="version-context-dropdown">
                    <ListboxInput value={props.versionContext} onChange={updateValue}>
                        {({ isExpanded }) => (
                            <>
                                <ListboxButton className="form-control">
                                    <span>{props.versionContext || 'Select context'}</span>
                                </ListboxButton>
                                <ListboxPopover
                                    className={classNames('version-context-dropdown__popover dropdown-menu', {
                                        show: isExpanded,
                                    })}
                                >
                                    <div className="version-context-dropdown__info">
                                        <span>Select version context</span>
                                        <div className="card">About version contexts</div>
                                    </div>
                                    <ListboxList className="version-context-dropdown__list">
                                        <ListboxOption
                                            disabled={true}
                                            value="title"
                                            className="version-context-dropdown__option version-context-dropdown__title"
                                        >
                                            <VersionContextInfoRow name="Name" description="Description" />
                                        </ListboxOption>
                                        {!isErrorLike(versionContexts) &&
                                            versionContexts
                                                .sort((a, b) => (a.name > b.name ? 1 : -1))
                                                .map(versionContext => (
                                                    <ListboxOption
                                                        key={versionContext.name}
                                                        value={versionContext.name}
                                                        label={versionContext.name}
                                                        className="version-context-dropdown__option"
                                                    >
                                                        <VersionContextInfoRow
                                                            name={versionContext.name}
                                                            description={versionContext.description}
                                                        />
                                                    </ListboxOption>
                                                ))}
                                    </ListboxList>
                                </ListboxPopover>
                            </>
                        )}
                    </ListboxInput>
                </div>
            ) : null}
        </>
    )
}

const VersionContextInfoRow: React.FunctionComponent<{ name: string; description: string }> = ({
    name,
    description,
}) => (
    <>
        <span />
        <span className="version-context-dropdown__option-name">{name}</span>
        <span className="version-context-dropdown__option-description">{description}</span>
        <span />
    </>
)
