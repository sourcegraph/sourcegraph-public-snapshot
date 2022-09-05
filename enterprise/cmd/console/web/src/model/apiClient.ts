import { Observable, of } from 'rxjs'

import { ConsoleAnonymousData, ConsoleData, ConsoleUserData } from './data'

const SAMPLE_ANONYMOUS_DATA: ConsoleAnonymousData = { user: null }

const SAMPLE_USER_DATA: ConsoleUserData = {
    user: {
        email: 'alice@example.com',
    },
    instances: [
        { id: 'c-f8a7d6ebfa8374ace', title: 'Sourcegraph', url: 'https://sourcegraph.sourcegraph.com' },
        { id: 'c-c9b7a7e739a6cd7cb', title: 'Acme Corp', url: 'https://acme.sourcegraph.com' },
        { id: 'c-f73a63765ac839a83', title: 'Initech', url: 'https://initech.sourcegraph.com' },
    ],
}

const DUMMY_USER = localStorage.getItem('signedIn') !== null

export function newAPIClient(): APIClient {
    return {
        getData: (): Observable<ConsoleData> => of(DUMMY_USER ? SAMPLE_USER_DATA : SAMPLE_ANONYMOUS_DATA),
    }
}

export interface APIClient {
    getData: () => Observable<ConsoleData>
}
