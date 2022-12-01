import React, { ReactNode, useCallback, useEffect, useState } from 'react'

import { mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import { RouteComponentProps } from 'react-router'
import { of } from 'rxjs'
import { delay, map } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useQuery } from '@sourcegraph/http-client/src'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Button,
    Code,
    Container,
    Icon,
    Link,
    LoadingSpinner,
    PageHeader,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Position,
    Text,
    Tooltip,
} from '@sourcegraph/wildcard'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { Timestamp } from '../components/time/Timestamp'
import { SlowRequestsResult, SlowRequestsVariables } from '../graphql-operations'

import { SLOW_REQUESTS, SLOW_REQUESTS_PAGE_POLL_INTERVAL } from './backend'

import styles from './SiteAdminSlowRequestsPage.module.scss'

export interface SiteAdminSlowRequestsPageProps extends RouteComponentProps, TelemetryProps {
    now?: () => Date
}

export type SlowRequest = SlowRequestsResult['slowRequests']['nodes'][0]

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filters',
        label: 'Filter',
        type: 'select',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all requests',
                args: {},
            },
            {
                label: 'Failed',
                value: 'failed',
                tooltip: 'Show only failed requests',
                args: { failed: true },
            },
            {
                label: 'Successful',
                value: 'successful',
                tooltip: 'Show only successful requests',
                args: { failed: false },
            },
        ],
    }
]

export const SiteAdminSlowRequestsPage: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminSlowRequestsPageProps>
> = ({ history, telemetryService }) => {
    const [items, setItems] = useState<SlowRequest[]>([])

    useEffect(() => {
        telemetryService.logPageView('SiteAdminSlowRequests')
    }, [telemetryService])

    var lastId = items[items.length - 1]?.id ?? null
    const { data, loading, error, stopPolling, refetch, startPolling } = useQuery<
        SlowRequestsResult,
        SlowRequestsVariables
    >(SLOW_REQUESTS, {
        variables: { after: lastId },
        pollInterval: SLOW_REQUESTS_PAGE_POLL_INTERVAL,
    })

    useEffect(() => {
        if (data?.slowRequests?.nodes?.length && (!lastId || data?.slowRequests.nodes[0].id > lastId)) {
            const newItems = items
                .concat(...data.slowRequests.nodes)
                .slice(
                    Math.max(
                        items.length +
                            data.slowRequests.nodes.length -
                            50,
                        0
                    )
                )
            console.log(newItems[newItems.length - 1]?.id)
            console.log(lastId)
            lastId = newItems[newItems.length - 1]?.id
            // Workaround for https://github.com/apollographql/apollo-client/issues/3053 to update the variables.
            // Weirdly enough, we don't need to wait for refetch() to complete before restarting polling.
            // See http://www.petecorey.com/blog/2019/09/23/apollo-quirks-polling-after-refetching-with-new-variables/
            stopPolling()
            setItems(newItems)
            refetch({ after: newItems[newItems.length - 1]?.id ?? null })
                .then(() => {})
                .catch(() => {})
            startPolling(SLOW_REQUESTS_PAGE_POLL_INTERVAL)
        }
    }, [data, lastId, items, refetch, startPolling, stopPolling])

    const querySlowRequests = useCallback(
        (args: FilteredConnectionQueryArguments & { failed?: boolean }) =>
            of([...items].reverse()).pipe(
                delay(200), // Without this, FilteredConnection will get into an infinite loop. :facepalm:
                map(items => {
                    const filtered = items?.filter(
                        request =>
                            (!args.query || matchesString(request, args.query)) && (args.failed !== !isSuccessful(request))
                    )
                    return {
                        nodes: filtered ?? [],
                        totalCount: (filtered ?? []).length,
                    }
                })
            ),
        [items]
    )

    return (
        <div className="site-admin-migrations-page">
            <PageTitle title="Slow requests - Admin" />
            <PageHeader
                path={[{ text: 'Slow requests' }]}
                headingElement="h2"
                description={
                    <>
                        This is the log of recent slow GraphQL requests received by the Sourcegraph instance. Handy for seeing
                        what's happening between clients and our API.{' '}
                        <strong>The list updates every five seconds.</strong>
                    </>
                }
                className="mb-3"
            />

            <Container className="mb-3">
                {error && !loading && <ErrorAlert error={error} />}
                {loading && !error && <LoadingSpinner />}
                {50 ? (
                    <FilteredConnection<SlowRequest>
                        className="mb-0"
                        listComponent="div"
                        listClassName={classNames('list-group mb-3', styles.requestsGrid)}
                        noun="request"
                        pluralNoun="requests"
                        queryConnection={querySlowRequests}
                        nodeComponent={MigrationNode}
                        filters={filters}
                        history={history}
                        location={history.location}
                    />
                ) : (
                    <>
                        <Text>Slow request logging is currently disabled.</Text>
                        <Text>
                            Set `slowRequestLogLimit` to a non-zero value in your{' '}
                            <Link to="/site-admin/configuration">site config</Link> to enable it.
                        </Text>
                    </>
                )}
            </Container>
        </div>
    )
}

const MigrationNode: React.FunctionComponent<{ node: React.PropsWithChildren<SlowRequest> }> = ({ node }) => {
    const roundedSecond = Math.round((node.duration + Number.EPSILON) * 100) / 100
    const [copied, setCopied] = useState(false)

    const copyToClipboard = (text: string): void => {
        copy(text)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
    }

    return (
        <React.Fragment key={node.id}>
            <span className={styles.separator} />
            <div className="flex-bounded">
                <Tooltip content="Duration">
                  <span>{roundedSecond.toFixed(2)} second{roundedSecond === 1 ? '' : 's'}</span>
                </Tooltip>
            </div>
            <div>
                <Tooltip content="Request Name">
                        <strong>{node.name}</strong>
                </Tooltip>
            </div>
            <div>
                <Tooltip content={"Repo Name: " + node.repoName}>
                        <span>{shortenRepoName(node.repoName)}</span>
                </Tooltip>
            </div>
            <div>
              <Tooltip content={'Filepath: ' + node.filepath}>
                <code>{ellipsis(node.filepath)}</code>
              </Tooltip>
            </div>
            <div>
                <Timestamp date={node.start} noAbout={true} />
            </div>
            <div className={classNames('d-flex flex-row')}>
                <div>
                    <Tooltip content={copied ? 'cURL command copied' : 'Copy cURL command (remember to edit ACCESS_TOKEN)'}>
                        <Button className="ml-2" onClick={() => copyToClipboard(buildCurlCommand(node))}>
                            cURL
                        </Button>
                    </Tooltip>
                </div>
                <SimplePopover label="More info">
                    <small className={styles.moreInfo}>
                        <Text>
                            <strong>Name: </strong>
                            {node.name}
                        </Text>
                        <Text>
                            <strong>UserID: </strong>
                            {node.userId}
                        </Text>
                        <Text>
                            <strong>Date/time started: </strong>
                            <Timestamp date={node.start} preferAbsolute={true} noAbout={true} />
                        </Text>
                        <Text>
                            <strong>Duration: </strong>
                            {roundedSecond.toFixed(2)} second{roundedSecond === 1 ? '' : 's'}
                        </Text>
                        <Text>
                          <strong>Errors: </strong> 
                          {node.errors.length > 0 ? node.errors : 'none'}
                        </Text>
                        <Text>
                          <strong>Variables: </strong>
                          <pre>{node.variables}</pre>
                        </Text>
                        <Text> 
                          <strong>Query: </strong>
                          <pre>{node.query}</pre>
                        </Text>
                    </small>
                </SimplePopover>
            </div>
        </React.Fragment>
    )
}

const SimplePopover: React.FunctionComponent<{ label: string; children: ReactNode }> = ({ label, children }) => {
    const [isOpen, setIsOpen] = useState(false)
    const handleOpenChange = useCallback(({ isOpen }: { isOpen: boolean }) => setIsOpen(isOpen), [setIsOpen])
    return (
        <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
            <PopoverTrigger as={Button} variant="secondary" outline={true}>
                <small>{label}</small>
                <Icon aria-label="Show details" svgPath={mdiChevronDown} />
            </PopoverTrigger>
            <PopoverContent position={Position.bottom} focusLocked={false}>
                <div className="p-2">{children}</div>
            </PopoverContent>
        </Popover>
    )
}

function isSuccessful(request: SlowRequest): boolean {
  return request.errors.length == 0
}

function matchesString(request: SlowRequest, query: string): boolean {
    const lQuery = query.toLowerCase()
    return (
        request.name.toLowerCase().includes(lQuery) ||
        request.repoName.toLowerCase().includes(lQuery) ||
        request.errors.some(error => error.toLowerCase().includes(lQuery), false)
    )
}

function ellipsis(str: string): string { 
  if (str.length <= 60) {
    return str
  } else {
    return '...' + str.slice(str.length - 50, str.length)
  }
}

function shortenRepoName(repoName: string): string { 
  return repoName.split('/').at(-1) || "" // TODO
}

function buildCurlCommand(request: SlowRequest): string {
  const headers = `-H 'Authorization: token YOUR_TOKEN'`
  const body = (`{"query": "${request.query}", "variables": ${request.variables}}`).replaceAll('\n', '')
  return `curl -X POST ${headers} -d '${body}' '${window.location.protocol}//${window.location.host}/.api/graphql?${request.name}'`
}
