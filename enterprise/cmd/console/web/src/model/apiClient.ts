import { Observable, of } from 'rxjs'

import { ConsoleData, ConsoleUserData } from './data'

const SAMPLE_DATA: ConsoleUserData = {
    user: {
        email: 'alice@example.com',
    },
    instances: [
        { id: 'c-f8a7d6ebfa8374ace', title: 'Sourcegraph', url: 'https://sourcegraph.sourcegraph.com' },
        { id: 'c-c9b7a7e739a6cd7cb', title: 'Acme Corp', url: 'https://acme.sourcegraph.com' },
        { id: 'c-f73a63765ac839a83', title: 'Initech', url: 'https://initech.sourcegraph.com' },
    ],
}

export function newAPIClient(): APIClient {
    return {
        getData: (): Observable<ConsoleData> => of(SAMPLE_DATA),
    }
}

export interface APIClient {
    getData: () => Observable<ConsoleData>
}
