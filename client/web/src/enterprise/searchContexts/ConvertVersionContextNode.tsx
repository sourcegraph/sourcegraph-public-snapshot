import classNames from 'classnames'
import React, { useCallback, useEffect, useState } from 'react'
import { merge, Observable, of, Subject } from 'rxjs'
import { catchError, delay, mergeMap, switchMap } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useEventObservable } from '@sourcegraph/shared/src/util/useObservable'

import styles from './ConvertVersionContextNode.module.scss'
import { ConvertVersionContextsPageProps } from './ConvertVersionContextsPage'

export interface ConvertVersionContextNodeProps
    extends Pick<
        ConvertVersionContextsPageProps,
        'convertVersionContextToSearchContext' | 'isSearchContextSpecAvailable'
    > {
    name: string
    searchContextSpec: string
    isConvertedUpdates: Subject<void>
}

const LOADING = 'LOADING' as const

export const ConvertVersionContextNode: React.FunctionComponent<ConvertVersionContextNodeProps> = ({
    name,
    searchContextSpec,
    isConvertedUpdates,
    convertVersionContextToSearchContext,
    isSearchContextSpecAvailable,
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
            [convertVersionContextToSearchContext, name]
        )
    )

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
    }, [isSearchContextSpecAvailable, isConvertedUpdates, searchContextSpec])

    useEffect(() => isConvertedUpdates.next(), [isConvertedUpdates])

    return (
        <div
            data-testid="convert-version-context-node"
            className={classNames(
                'd-flex justify-content-between align-items-center flex-row',
                styles.convertVersionContextNode
            )}
        >
            <div>{name}</div>
            {(convertOrError === LOADING || isConverted === LOADING) && <LoadingSpinner />}
            {isConverted === false && !convertOrError && (
                <button
                    type="button"
                    className="btn btn-sm btn-primary test-convert-version-context-btn"
                    onClick={convert}
                >
                    Convert
                </button>
            )}
            {!convertOrError && isConverted === true && (
                <div className="text-muted test-converted-context">Converted</div>
            )}
            {isErrorLike(convertOrError) && (
                <div className="text-danger">
                    <strong>Error:</strong> {convertOrError.message}
                </div>
            )}
            {convertOrError && convertOrError !== LOADING && !isErrorLike(convertOrError) && (
                <div className="text-success">Version context successfully converted.</div>
            )}
        </div>
    )
}
