import React, { useEffect, useMemo } from 'react'

// export const SymbolsPage: React.FunctionComponent<Props> = ({ repo, resolvedRev, viewOptions, history, ...props }) => {
//     useEffect(() => {
//         eventLogger.logViewEvent('Symbols')
//     }, [])

//     const data = useObservable(
//         useMemo(
//             () =>
//                 queryRepositorySymbols({
//                     repo: repo.id,
//                     commitID: resolvedRev.commitID,
//                     path: '.',
//                     filters: viewOptions,
//                 }),
//             [repo.id, resolvedRev.commitID, viewOptions]
//         )
//     )

//     return data ? <ContainerSymbolsList symbols={data} history={history} /> : <LoadingSpinner className="m-3" />
// }

export interface SymbolsRouteProps {}

export const SymbolsPage: React.FunctionComponent<SymbolsRouteProps> = () => {
    return <div>These are the symbols</div>
}
