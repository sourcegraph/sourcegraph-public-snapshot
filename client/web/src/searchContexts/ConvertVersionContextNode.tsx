import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { convertVersionContextToSearchContext } from '../search/backend'
import { useEventObservable } from '@sourcegraph/shared/src/util/useObservable'
import { catchError, delay, mergeMap, switchMap } from 'rxjs/operators'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { merge, Observable, of, Subject } from 'rxjs'
import { isSearchContextSpecAvailable } from '../search'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

export interface ConvertVersionContextNodeProps {
    name: string
    isConvertedUpdates: Subject<void>
}

const LOADING = 'LOADING' as const

const versionContextNameToSearchContextSpecRegExp = /\s+/g

export const ConvertVersionContextNode: React.FunctionComponent<ConvertVersionContextNodeProps> = ({
    name,
    isConvertedUpdates,
}) => {
    const [convert, convertOrError] = useEventObservable(
        useCallback(
            (event: Observable<React.MouseEvent>) =>
                event.pipe(
                    mergeMap(() => {
                        const conversion = convertVersionContextToSearchContext(name).pipe(
                            catchError(error => [asError(error)])
                        )
                        return merge(of(LOADING), conversion.pipe(delay(500)))
                    })
                ),
            [name]
        )
    )

    const searchContextSpec = useMemo(() => name.replaceAll(versionContextNameToSearchContextSpecRegExp, '_'), [name])

    const [isConverted, setIsConverted] = useState<boolean | typeof LOADING>(false)
    useEffect(() => {
        const subscription = isConvertedUpdates
            .pipe(
                switchMap(() =>
                    merge(of(LOADING), isSearchContextSpecAvailable(searchContextSpec, true).pipe(delay(250)))
                )
            )
            .subscribe(result => setIsConverted(result))

        return () => subscription.unsubscribe()
    }, [isConvertedUpdates, searchContextSpec])

    useEffect(() => isConvertedUpdates.next(), [isConvertedUpdates])

    return (
        <div className="convert-version-context-node card mb-1 p-3">
            <div>{name}</div>
            {(convertOrError === LOADING || isConverted === LOADING) && <LoadingSpinner />}
            {isConverted === false && !convertOrError && (
                <button type="button" className="btn btn-primary test-convert-version-context-btn" onClick={convert}>
                    Convert
                </button>
            )}
            {!convertOrError && isConverted === true && (
                <div className="text-muted test-converted-context">Converted</div>
            )}
            {isErrorLike(convertOrError) && (
                <div className="alert-danger mt-1 p-2">
                    <strong>Error:</strong> {convertOrError.message}
                </div>
            )}
            {convertOrError && convertOrError !== LOADING && !isErrorLike(convertOrError) && (
                <div className="alert-success mt-1 p-2">Version context successfully converted.</div>
            )}
        </div>
    )
}
