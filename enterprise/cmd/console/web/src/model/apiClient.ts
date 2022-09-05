import { Observable, of } from 'rxjs'

import { ConsoleAnonymousData, ConsoleData, ConsoleUserData } from './data'

const SAMPLE_ANONYMOUS_DATA: ConsoleAnonymousData = { user: null }

const SAMPLE_USER_DATA: ConsoleUserData = {
    user: {
        email: 'alice@example.com',
    },
    instances: [
        {
            id: 'c-f8a7d6ebfa8374ace',
            ownerName: 'Carol Lopez',
            ownerEmail: 'carol@sourcegraph.com',
            url: 'https://sourcegraph.sourcegraph.com',
            viewerIsOwner: true,
            viewerIsOrganizationMember: true,
            viewerCanMaybeSignIn: true,
            status: 'ready',
        },
        {
            id: 'c-c9b7a7e739a6cd7cb',
            ownerName: 'Alice Smith',
            ownerEmail: 'alice@acme-corp.com',
            url: 'https://acme.sourcegraph.com',
            viewerIsOwner: false,
            viewerIsOrganizationMember: true,
            viewerCanMaybeSignIn: true,
            status: 'ready',
        },
        {
            id: 'c-f73a63765ac839a83',
            ownerName: 'Fangfang Zhao',
            ownerEmail: 'ffz@example.com',
            url: 'https://initech.sourcegraph.com',
            viewerIsOwner: false,
            viewerIsOrganizationMember: true,
            viewerCanMaybeSignIn: true,
            status: 'ready',
        },
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
