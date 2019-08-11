import { useEffect, useMemo, useState } from 'react'
import { combineLatest, merge, Observable, of, Subject } from 'rxjs'
import { delay, map, switchMap, takeUntil, tap, throttleTime, withLatestFrom } from 'rxjs/operators'
import { fromDiagnostic } from '../../../../../../../shared/src/api/types/diagnostic'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../../shared/src/util/errors'
import { propertyIsDefined } from '../../../../../../../shared/src/util/types'
import { queryGraphQL } from '../../../../../backend/graphql'
import {
    diffStatFieldsFragment,
    fileDiffHunkRangeFieldsFragment,
} from '../../../../../repo/compare/RepositoryCompareDiffPage'
import { RuleDefinition } from '../../../../rules/types'
import {
    DiagnosticInfo,
    diagnosticQueryMatcher,
    getCodeActions,
    getDiagnosticInfos,
} from '../../../../threadsOLD/detail/backend'
import { computeDiff, FileDiff } from '../../../../threadsOLD/detail/changes/computeDiff'

export const CampaignPreviewFragment = gql`
    fragment CampaignPreviewFragment on CampaignPreview {
        name
        threads {
            nodes {
                __typename
                ... on ThreadPreview {
                    repository {
                        id
                        name
                        url
                    }
                    title
                }
            }
            totalCount
        }
        diagnostics {
            nodes {
                type
                data
            }
            totalCount
        }
        repositoryComparisons {
            baseRepository {
                id
                name
                url
            }
            headRepository {
                id
                name
                url
            }
            range {
                expr
                baseRevSpec {
                    object {
                        oid
                    }
                    expr
                }
                headRevSpec {
                    expr
                }
            }
            fileDiffs {
                nodes {
                    oldPath
                    newPath
                    hunks {
                        oldRange {
                            ...FileDiffHunkRangeFields
                        }
                        oldNoNewlineAt
                        newRange {
                            ...FileDiffHunkRangeFields
                        }
                        section
                        body
                    }
                    stat {
                        ...DiffStatFields
                    }
                    internalID
                }
                totalCount
                pageInfo {
                    hasNextPage
                }
                diffStat {
                    ...DiffStatFields
                }
            }
        }
    }
    ${fileDiffHunkRangeFieldsFragment}
    ${diffStatFieldsFragment}
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.ICampaignPreview | ErrorLike

interface DiagnosticsAndFileDiffs {
    diagnostics: DiagnosticInfo[]
    fileDiffs: Pick<FileDiff, 'patchWithFullURIs'>[]
}

const getDiagnosticsAndFileDiffs = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    rule: RuleDefinition
): Observable<DiagnosticsAndFileDiffs> => {
    if (rule.type !== 'DiagnosticRule') {
        return of({ diagnostics: [], fileDiffs: [] })
    }
    const matchesQuery = diagnosticQueryMatcher(rule.query)
    return getDiagnosticInfos(extensionsController, 'packageJsonDependency').pipe(
        map(diagnostics => diagnostics.filter(matchesQuery)),
        switchMap(diagnostics =>
            diagnostics.length > 0
                ? combineLatest(
                      diagnostics.map(d =>
                          getCodeActions({
                              diagnostic: d,
                              extensionsController,
                          }).pipe(
                              map(actions => ({
                                  diagnostic: d,
                                  action:
                                      rule.action !== undefined
                                          ? actions
                                                .filter(propertyIsDefined('computeEdit'))
                                                .find(a => a.computeEdit && a.computeEdit.command === rule.action)
                                          : undefined,
                              }))
                          )
                      )
                  )
                : of([])
        ),
        switchMap(async diagnosticsAndActions => {
            const fileDiffs = await computeDiff(
                extensionsController,
                diagnosticsAndActions
                    .filter(propertyIsDefined('action'))
                    .map(d => ({
                        actionEditCommand: d.action.computeEdit,
                        diagnostic: fromDiagnostic(d.diagnostic),
                    }))
                    .filter(propertyIsDefined('actionEditCommand'))
            )
            return { diagnostics: diagnosticsAndActions.map(({ diagnostic }) => diagnostic), fileDiffs }
        })
    )
}

/**
 * A React hook that observes a campaign preview queried from the GraphQL API.
 */
export const useCampaignPreview = (
    { extensionsController }: ExtensionsControllerProps,
    input: GQL.ICreateCampaignInput
): [Result, boolean] => {
    const inputSubject = useMemo(() => new Subject<GQL.ICreateCampaignInput>(), [])
    const [isLoading, setIsLoading] = useState(true)
    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => inputSubject.next(input), [input, inputSubject])
    useEffect(() => {
        const subscription = inputSubject
            .pipe(
                throttleTime(1000, undefined, { leading: true, trailing: true }),
                switchMap(input => {
                    const result = (input.rules && input.rules.length > 0
                        ? combineLatest(
                              (input.rules || []).map(rule => {
                                  const def: RuleDefinition = JSON.parse(rule.definition)
                                  return getDiagnosticsAndFileDiffs(extensionsController, def)
                              })
                          )
                        : of<DiagnosticsAndFileDiffs[]>([{ diagnostics: [], fileDiffs: [] }])
                    ).pipe(
                        map(results => {
                            const combined: DiagnosticsAndFileDiffs = {
                                diagnostics: results.flatMap(r => r.diagnostics),
                                fileDiffs: results.flatMap(r => r.fileDiffs),
                            }
                            return combined
                        }),
                        switchMap(({ diagnostics, fileDiffs }) =>
                            queryGraphQL(
                                gql`
                                    query CampaignPreview($input: CampaignPreviewInput!) {
                                        campaignPreview(input: $input) {
                                            ...CampaignPreviewFragment
                                        }
                                    }
                                    ${CampaignPreviewFragment}
                                `,
                                // tslint:disable-next-line: no-object-literal-type-assertion
                                {
                                    input: {
                                        campaign: input,
                                        rawDiagnostics: diagnostics.map(d =>
                                            // tslint:disable-next-line: no-object-literal-type-assertion
                                            JSON.stringify({
                                                __typename: 'Diagnostic',
                                                type: d.type,
                                                data: d,
                                            } as GQL.IDiagnostic)
                                        ),
                                        rawFileDiffs: fileDiffs.map(({ patchWithFullURIs }) => patchWithFullURIs),
                                    },
                                } as GQL.ICampaignPreviewOnQueryArguments
                            ).pipe(
                                map(dataOrThrowErrors),
                                map(data => data.campaignPreview),
                                tap(data => {
                                    // TODO!(sqs) hack, compensate for the RepositoryComparison head not existing
                                    for (const c of data.repositoryComparisons) {
                                        c.range.headRevSpec.object = { oid: '' } as any
                                        for (const d of c.fileDiffs.nodes) {
                                            d.mostRelevantFile = { path: d.newPath, url: '' } as any
                                        }
                                    }
                                })
                            )
                        )
                    )
                    return merge(result, of(LOADING).pipe(takeUntil(result)))
                })
            )
            .subscribe(
                result => {
                    setResult(prevResult => {
                        setIsLoading(result === LOADING)
                        // Reuse last non-error result while loading, to reduce UI jitter.
                        return result === LOADING && prevResult !== LOADING && !isErrorLike(prevResult)
                            ? prevResult
                            : result
                    })
                },
                err => {
                    setIsLoading(false)
                    setResult(asError(err))
                }
            )
        return () => subscription.unsubscribe()
    }, [extensionsController, inputSubject])
    return [result, isLoading]
}
