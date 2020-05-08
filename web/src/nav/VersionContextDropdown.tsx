import React, { useState, useMemo } from 'react'
import { Listbox, ListboxOption, ListboxInput, ListboxButton, ListboxPopover, ListboxList } from '@reach/listbox'
import { VersionContext } from '../schema/site.schema'
import { useObservable } from '../../../shared/src/util/useObservable'
import { fetchVersionContexts } from '../search/backend'
import { catchError } from 'rxjs/operators'
import { of } from 'rxjs'
import { ErrorLike, asError, isErrorLike } from '../../../shared/src/util/errors'
import { IVersionContext } from '../../../shared/src/graphql/schema'

interface Props {
    versionContext: string
    setVersionContext: (versionContext: string) => void
}
export const VersionContextDropdown: React.FunctionComponent<Props> = (props: Props) => {
    // const versionContexts: VersionContext[] | undefined =
    //     window.context.experimentalFeatures &&
    //     window.context.experimentalFeatures.versionContexts?.map((versionContext: VersionContext) => {
    //         versionContext.name
    //         versionContext.description
    //         [...versionContext.revisions]
    //     })
    const versionContexts = useObservable(
        useMemo(() => fetchVersionContexts().pipe(catchError(err => of<ErrorLike>(asError(err)))), [])
    )

    const updateValue = (newValue: string): void => {
        console.log('newvalue', newValue)
        props.setVersionContext(newValue)
    }

    return (
        <div className="px-3">
            <Listbox arrow={true} value={props.versionContext} onChange={updateValue}>
                <ListboxInput>
                    {({ isExpanded }) => (
                        <>
                            <ListboxButton />
                            <ListboxPopover>
                                <ListboxList>
                                    {!isErrorLike(versionContexts) &&
                                        versionContexts?.map((versionContext: IVersionContext) => (
                                            <ListboxOption key={versionContext.name} value={versionContext.name}>
                                                {versionContext.name}
                                            </ListboxOption>
                                        ))}
                                </ListboxList>
                            </ListboxPopover>
                        </>
                    )}
                </ListboxInput>
            </Listbox>
        </div>
    )
}
