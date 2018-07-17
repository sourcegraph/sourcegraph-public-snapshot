import * as assert from 'assert'
import { Observable } from 'rxjs'
import { first } from 'rxjs/operators'
import { Client } from '../client/client'
import { Controller } from './controller'
import { Environment } from './environment'

class TestController extends Controller {
    public get clients(): Observable<Client[]> & { value: Client[] } {
        let value!: Client[]
        super.clients
            .pipe(first())
            .subscribe(clients => (value = clients))
            .unsubscribe()
        return { ...super.clients, value } as Observable<Client[]> & { value: Client[] }
    }
}

const create = (environment?: Environment): TestController => {
    const controller = new TestController({
        createMessageTransports: async () => {
            throw new Error('connection is not used in unit test')
        },
    })
    if (environment) {
        controller.setEnvironment(environment)
    }
    return controller
}

const FIXTURE_ENVIRONMENT: Environment = {
    root: 'file:///',
    component: {
        document: { uri: 'file:///f', languageId: 'l' },
        selections: [],
        visibleRanges: [],
    },
    extensions: [{ id: 'x', settings: { merged: {} } }],
}

describe('Controller', () => {
    it('creates clients for the environment', () => {
        const controller = create(FIXTURE_ENVIRONMENT)
        assert.deepStrictEqual(
            controller.clients.value.map(client => ({ id: client.id, root: client.clientOptions.root })),
            [{ id: 'x', root: 'file:///' }]
        )
    })

    it('creates no clients if the environment needs none', () => {
        assert.deepStrictEqual(create({ ...FIXTURE_ENVIRONMENT, root: null }).clients.value, [])
    })
})
