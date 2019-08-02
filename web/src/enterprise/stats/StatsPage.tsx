import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import ChartLineIcon from 'mdi-react/ChartLineIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import React, { useCallback, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { numberWithCommas } from '../../../../shared/src/util/strings'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { Form } from '../../components/Form'
import { useStatisticsForSearchResults } from './useStatisticsForSearchResults'

interface Props extends RouteComponentProps<{}> {}

const LOADING = 'loading' as const

export const StatsPage: React.FunctionComponent<Props> = ({ location, history }) => {
    const query = new URLSearchParams(location.search).get('q') || ''
    const [uncommittedQuery, setUncommittedQuery] = useState(query)
    const onUncommittedQueryChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(e => {
        setUncommittedQuery(e.currentTarget.value)
    }, [])
    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        e => {
            e.preventDefault()
            history.push({ ...location, search: new URLSearchParams({ q: uncommittedQuery }).toString() })
        },
        [history, location, uncommittedQuery]
    )

    const stats = useStatisticsForSearchResults(query + ' count:99999999')

    const urlToSearchWithExtraQuery = useCallback(
        (extraQuery: string) => `/search?${buildSearchURLQuery(`${query} ${extraQuery}`)}`,
        [query]
    )

    return (
        <div className="stats-page container mt-4">
            <header>
                <h2 className="d-flex align-items-center">
                    <ChartLineIcon className="icon-inline mr-2" /> Statistics
                </h2>
            </header>
            <Form onSubmit={onSubmit} className="form">
                <div className="form-group d-flex align-items-stretch">
                    <label htmlFor="stats-page__query" className="mr-2 mt-3 pt-1 pr-1" aria-label="Query">
                        <SearchIcon className="icon-inline" />
                    </label>
                    <input
                        id="stats-page__query"
                        className="form-control mr-2 flex-1"
                        style={{ maxWidth: '50%' }}
                        type="search"
                        placeholder="Filter statistics..."
                        value={uncommittedQuery}
                        onChange={onUncommittedQueryChange}
                        autoCapitalize="off"
                        spellCheck={false}
                        autoCorrect="off"
                        autoComplete="off"
                    />
                    {uncommittedQuery !== query && (
                        <button type="submit" className="btn btn-primary">
                            Update
                        </button>
                    )}
                </div>
            </Form>
            <hr className="my-3" />
            {stats === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(stats) ? (
                <div className="alert alert-danger">{stats.message}</div>
            ) : (
                <div className="border" style={{ maxWidth: '20rem', maxHeight: '50vh', overflowY: 'auto' }}>
                    <table className="table">
                        <thead>
                            <tr>
                                <th>Language</th>
                                <th>Lines</th>
                            </tr>
                        </thead>
                        <tbody>
                            {stats.languages.map(({ name, totalBytes }) => (
                                <tr key={name}>
                                    <td>
                                        <Link to={urlToSearchWithExtraQuery(`lang:${name.toLowerCase()}`)}>{name}</Link>
                                    </td>
                                    <td>{numberWithCommas(Math.ceil(totalBytes / 31))}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    )
}
