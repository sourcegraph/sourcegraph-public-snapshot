import React, { useCallback, useMemo, useState } from 'react'
import { combineLatest, concat, from, Observable, of, Subject } from 'rxjs'
import { catchError, concatMap, delay, map, mergeMap, reduce, startWith, tap, toArray } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { VirtualList } from '@sourcegraph/shared/src/components/VirtualList'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useEventObservable, useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'

import { VersionContext } from '../../schema/site.schema'
import { SearchContextProps } from '../../search'

import { ConvertVersionContextNode } from './ConvertVersionContextNode'
import styles from './ConvertVersionContextsPage.module.scss'

export interface ConvertVersionContextsPageProps
    extends Pick<SearchContextProps, 'convertVersionContextToSearchContext' | 'isSearchContextSpecAvailable'> {
    availableVersionContexts: VersionContext[] | undefined
}

const initialItemsToShow = 15
const incrementalItemsToShow = 10
const LOADING = 'LOADING' as const

const versionContextNameToSearchContextSpecRegExp = /\s+/g

export const ConvertVersionContextsPage: React.FunctionComponent<ConvertVersionContextsPageProps> = ({
    availableVersionContexts,
    convertVersionContextToSearchContext,
    isSearchContextSpecAvailable,
}) => {
    const itemKey = useCallback((item: VersionContext): string => item.name, [])

    const versionContexts = useObservable(
        useMemo(() => {
            if (!availableVersionContexts) {
                return of([])
            }
            return from(availableVersionContexts).pipe(
                concatMap(versionContext => {
                    const searchContextSpec = versionContext.name.replace(
                        versionContextNameToSearchContextSpecRegExp,
                        '_'
                    )
                    return combineLatest([
                        of({
                            ...versionContext,
                            searchContextSpec,
                            isConvertedUpdates: new Subject<void>(),
                        }),
                        isSearchContextSpecAvailable(searchContextSpec),
                    ])
                }),
                map(([versionContext, isConverted]) => ({ ...versionContext, isConverted })),
                toArray(),
                startWith(LOADING)
            )
        }, [isSearchContextSpecAvailable, availableVersionContexts])
    )

    // Sort unconverted version contexts to the front of the array
    const sortedVersionContexts = useMemo(
        () =>
            versionContexts && versionContexts !== LOADING
                ? versionContexts.sort((a, b) => Number(a.isConverted) - Number(b.isConverted))
                : [],
        [versionContexts]
    )

    const allVersionContextsConverted = useMemo(
        () => sortedVersionContexts.every(versionContext => versionContext.isConverted),
        [sortedVersionContexts]
    )

    const renderResult = useCallback(
        (item: VersionContext & { isConvertedUpdates: Subject<void>; searchContextSpec: string }): JSX.Element => (
            <ConvertVersionContextNode
                name={item.name}
                searchContextSpec={item.searchContextSpec}
                isConvertedUpdates={item.isConvertedUpdates}
                isSearchContextSpecAvailable={isSearchContextSpecAvailable}
                convertVersionContextToSearchContext={convertVersionContextToSearchContext}
            />
        ),
        [isSearchContextSpecAvailable, convertVersionContextToSearchContext]
    )

    const [itemsToShow, setItemsToShow] = useState(initialItemsToShow)
    const onBottomHit = useCallback(() => {
        setItemsToShow(items => Math.min(sortedVersionContexts.length || 0, items + incrementalItemsToShow))
    }, [sortedVersionContexts])

    const [convertAll, convertAllResult] = useEventObservable(
        useCallback(
            (event: Observable<React.MouseEvent>) =>
                event.pipe(
                    mergeMap(() => {
                        const convertAll = from(sortedVersionContexts).pipe(
                            mergeMap(({ name, isConvertedUpdates }) =>
                                convertVersionContextToSearchContext(name).pipe(
                                    tap(() => isConvertedUpdates.next()),
                                    catchError(error => [asError(error)])
                                )
                            ),
                            map(result => (isErrorLike(result) ? 0 : 1)),
                            reduce((accumulator, result) => accumulator + result, 0)
                        )
                        return concat(of(LOADING), convertAll.pipe(delay(500)))
                    })
                ),
            [convertVersionContextToSearchContext, sortedVersionContexts]
        )
    )

    return (
        <div className="w-100">
            <Page>
                <div className="container col-8">
                    <PageTitle title="Convert version contexts" />
                    <div>
                        <Link to="/contexts">
                            « <span className={styles.backLabel}>Back</span>
                        </Link>
                        <div className="page-header d-flex flex-wrap align-items-center mt-2">
                            <h2 className="flex-grow-1">Convert version contexts</h2>
                        </div>
                        <div className="text-muted">
                            Convert existing version contexts defined in site config into search contexts.{' '}
                            <a
                                href="https://docs.sourcegraph.com/code_search/explanations/features#search-contexts"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                Learn more
                            </a>
                        </div>
                        <div className="d-flex flex-row justify-content-between align-items-center mt-4">
                            <h3>Available version contexts</h3>
                            <button
                                type="button"
                                className="btn btn-outline-primary test-convert-all-search-contexts-btn"
                                onClick={convertAll}
                                disabled={
                                    allVersionContextsConverted ||
                                    convertAllResult === LOADING ||
                                    typeof convertAllResult !== 'undefined'
                                }
                            >
                                {convertAllResult === LOADING ? 'Converting All...' : 'Convert All'}
                            </button>
                        </div>
                        {typeof convertAllResult !== 'undefined' &&
                            convertAllResult !== LOADING &&
                            (convertAllResult === 0 ? (
                                <div className="alert alert-info mt-3">No version contexts to convert.</div>
                            ) : (
                                <div className="alert alert-success test-convert-all-search-contexts-success mt-3">
                                    Sucessfully converted <strong>{convertAllResult}</strong> version contexts into
                                    search contexts.
                                </div>
                            ))}
                        <hr className="mt-3 mb-0" />
                        {versionContexts && versionContexts === LOADING && (
                            <div className="d-flex justify-content-center mt-3">
                                <LoadingSpinner />
                            </div>
                        )}
                        {versionContexts && versionContexts !== LOADING && sortedVersionContexts.length > 0 && (
                            <VirtualList<
                                VersionContext & { isConvertedUpdates: Subject<void>; searchContextSpec: string }
                            >
                                itemsToShow={itemsToShow}
                                onShowMoreItems={onBottomHit}
                                items={sortedVersionContexts}
                                itemProps={undefined}
                                itemKey={itemKey}
                                renderItem={renderResult}
                            />
                        )}
                        {versionContexts && versionContexts !== LOADING && sortedVersionContexts.length === 0 && (
                            <div className="d-flex justify-content-center mt-3 text-muted">
                                No version contexts to convert.
                            </div>
                        )}
                    </div>
                </div>
            </Page>
        </div>
    )
}
