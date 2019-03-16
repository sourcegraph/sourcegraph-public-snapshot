import { Observable, of, throwError } from 'rxjs'
import { ErrorLike } from '../../../../shared/src/util/errors'
import { Connection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'

export interface Check {
    id: string
    title: string
    status: 'open' | 'closed' | 'disabled'
    labels: string[]

    commitID: string
    author: string
    count: number
    messageCount: number
    timeAgo: string
}

export const CHECKS: Check[] = [
    {
        id: 'w8aolhf',
        status: 'open',
        commitID: '0c3c511',
        author: 'ekonev',
        count: 217,
        title: 'Possible database files',
        messageCount: 5,
        timeAgo: '3 hours ago',
        labels: ['security', 'infosec'],
    },
    {
        id: 'ahvn4yqf',
        status: 'open',
        commitID: '8c3537f',
        author: 'jleiner',
        count: 73,
        title: 'Potential cryptographic private keys',
        messageCount: 0,
        timeAgo: '9 hours ago',
        labels: ['security', 'infosec'],
    },
    {
        id: 'ia8uh4gy',
        status: 'open',
        commitID: '17c630d',
        author: 'jleiner',
        count: 112,
        title: 'Potential API secret keys',
        messageCount: 3,
        timeAgo: '1 day ago',
        labels: ['security', 'infosec', 'noisy'],
    },
    {
        id: '8727gzq4',
        status: 'open',
        commitID: '910c03',
        author: 'blslevitsky',
        count: 2,
        title: 'New API consumers',
        messageCount: 1,
        timeAgo: '1 day ago',
        labels: ['tech-lead', 'services'],
    },
    {
        id: '239bn35a',
        status: 'open',
        commitID: 'af7b381',
        author: 'kting7',
        count: 3,
        title: 'New npm dependencies',
        messageCount: 7,
        timeAgo: '2 days ago',
        labels: ['security', 'appsec', 'build'],
    },
    {
        id: '76t3syua',
        status: 'open',
        commitID: '0cba83d',
        author: 'ziyang',
        count: 2,
        title: 'Untrusted publishers of npm dependencies',
        messageCount: 21,
        timeAgo: '2 days ago',
        labels: ['security', 'appsec', 'build'],
    },
    {
        id: 'dbfni7gmo',
        status: 'open',
        commitID: 'c8164ef',
        author: 'ffranksena',
        count: 8,
        title: 'Code with no owner',
        messageCount: 14,
        timeAgo: '3 days ago',
        labels: ['tech-lead'],
    },
    {
        id: '3wsdfty78',
        status: 'open',
        commitID: '83c713',
        author: 'aconnor93',
        count: 3,
        title: 'PRs waiting on review >24h SLA',
        messageCount: 0,
        timeAgo: '3 days ago',
        labels: ['tech-lead'],
    },
]

export const CHECKS_CONNECTION: Connection<Check> = {
    nodes: CHECKS,
    totalCount: CHECKS.length,
    pageInfo: { hasNextPage: false },
}

export function queryChecks(args: FilteredConnectionQueryArgs): Observable<Connection<Check>> {
    return of(CHECKS_CONNECTION)
}

export function queryCheck(id: Check['id']): Observable<Check | ErrorLike> {
    const check = CHECKS.find(check => check.id === id)
    return check ? of(check) : throwError(new Error('check not found'))
}
