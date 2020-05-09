import React, { useMemo } from 'react'
import { ListboxOption, ListboxInput, ListboxButton, ListboxPopover, ListboxList } from '@reach/listbox'
import { useObservable } from '../../../shared/src/util/useObservable'
import { fetchVersionContexts } from '../search/backend'
import { catchError } from 'rxjs/operators'
import { of } from 'rxjs'
import { ErrorLike, asError, isErrorLike } from '../../../shared/src/util/errors'
import { IVersionContext } from '../../../shared/src/graphql/schema'
import classNames from 'classnames'

interface Props {
    versionContext: string
    setVersionContext: (versionContext: string) => void
}
export const VersionContextDropdown: React.FunctionComponent<Props> = (props: Props) => {
    const versionContexts = useObservable(
        useMemo(() => fetchVersionContexts().pipe(catchError(err => of<ErrorLike>(asError(err)))), [])
    )

    const updateValue = (newValue: string): void => {
        props.setVersionContext(newValue)
    }

    return (
        <>
            {versionContexts ? (
                <div>
                    <ListboxInput onChange={updateValue}>
                        {({ valueLabel, isExpanded }) => (
                            <>
                                <ListboxButton>{valueLabel}</ListboxButton>
                                <ListboxPopover className={classNames('dropdown-menu', { show: isExpanded })}>
                                    <ListboxList>
                                        {!isErrorLike(versionContexts) &&
                                            versionContexts.map((versionContext: IVersionContext) => (
                                                <ListboxOption key={versionContext.name} value={versionContext.name}>
                                                    {versionContext.name}
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
