import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { truncate } from 'lodash'
import ChartLineIcon from 'mdi-react/ChartLineIcon'
import EmailOpenOutlineIcon from 'mdi-react/EmailOpenOutlineIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React, { useCallback, useState } from 'react'
import PieChart, { LabelProps } from 'react-minimal-pie-chart'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle, UncontrolledButtonDropdown } from 'reactstrap'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { numberWithCommas } from '../../../../shared/src/util/strings'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { Form } from '../../components/Form'
import { useStatisticsForSearchResults } from './useStatisticsForSearchResults'

interface Props extends RouteComponentProps<{}> {}

const LOADING = 'loading' as const

const COLORS = ['#278389', '#f16321', '#ff7700', '#651fff', '#0091ea', '#00c853', '#ffab00', '#ff3d00']

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
            <header className="d-flex align-items-center justify-content-between mb-3">
                <h2 className="d-flex align-items-center mb-0">
                    <ChartLineIcon className="icon-inline mr-2" /> Statistics
                </h2>
                <UncontrolledButtonDropdown>
                    <DropdownToggle caret={true} color="" className="btn-primary">
                        Publish
                    </DropdownToggle>
                    <DropdownMenu>
                        <DropdownItem>
                            <SlackIcon className="icon-inline" /> Slack channel...
                        </DropdownItem>
                        <DropdownItem>
                            <EmailOpenOutlineIcon className="icon-inline" /> Email...
                        </DropdownItem>
                    </DropdownMenu>
                </UncontrolledButtonDropdown>
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
                <>
                    <div className="mb-3">
                        {stats.languages.length > 0 ? (
                            <div className="d-flex border align-items-stretch">
                                <div className="flex-1 border-right" style={{ maxHeight: '50vh', overflowY: 'auto' }}>
                                    <table className="table mb-0">
                                        <thead>
                                            <tr>
                                                <th>Language</th>
                                                <th>Lines</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {stats.languages.map(({ name, totalBytes }, i) => (
                                                <tr key={name || i}>
                                                    <td>
                                                        <Link
                                                            to={urlToSearchWithExtraQuery(`lang:${name.toLowerCase()}`)}
                                                        >
                                                            {name}
                                                        </Link>
                                                    </td>
                                                    <td>{numberWithCommas(totalBytes)}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                                <PieChart
                                    data={stats.languages.slice(0, COLORS.length).map(({ name, totalBytes }, i) => ({
                                        title: name,
                                        value: totalBytes,
                                        color: COLORS[i % COLORS.length],
                                    }))}
                                    labelStyle={{ fillColor: 'white', fill: 'white', fontSize: '0.25rem' }}
                                    label={props => props.data[props.dataIndex].title}
                                    // label={(props: LabelProps) => (
                                    //     <span className="text-white">{props.data[props.dataIndex].title}</span>
                                    // )}
                                    className="flex-1 m-6 p-3"
                                    style={{ maxHeight: '22rem' }}
                                />
                            </div>
                        ) : (
                            <div className="text-muted p-2">No language statistics available</div>
                        )}
                    </div>
                    <div className="mb-3">
                        {stats.owners.length > 0 ? (
                            <div className="d-flex border align-items-stretch">
                                <div className="flex-1 border-right" style={{ maxHeight: '50vh', overflowY: 'auto' }}>
                                    <table className="table mb-0">
                                        <thead>
                                            <tr>
                                                <th>Owner</th>
                                                <th>Lines</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {stats.owners.map(({ owner, totalBytes }) => (
                                                <tr key={owner}>
                                                    <td>
                                                        <Link
                                                            to={urlToSearchWithExtraQuery(
                                                                `${
                                                                    query.includes('type:diff') ? '' : 'type:diff '
                                                                }author:${owner}`
                                                            )}
                                                        >
                                                            {owner}
                                                        </Link>
                                                    </td>
                                                    <td>{numberWithCommas(totalBytes)}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                                <PieChart
                                    data={stats.owners.slice(0, 5).map(({ owner, totalBytes }, i) => ({
                                        title: truncate(owner, 12),
                                        value: totalBytes,
                                        color: COLORS[i % COLORS.length],
                                    }))}
                                    labelStyle={{ fillColor: 'white', fill: 'white', fontSize: '0.25rem' }}
                                    label={props => props.data[props.dataIndex].title}
                                    // label={(props: LabelProps) => (
                                    //     <span className="text-white">{props.data[props.dataIndex].title}</span>
                                    // )}
                                    className="flex-1 m-6 p-3"
                                    style={{ maxHeight: '22rem' }}
                                />
                            </div>
                        ) : (
                            <div className="text-muted p-2">No ownership statistics available</div>
                        )}
                    </div>
                </>
            )}
        </div>
    )
}
