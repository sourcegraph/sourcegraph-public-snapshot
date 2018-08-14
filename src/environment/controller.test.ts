import * as assert from 'assert'
import { Observable } from 'rxjs'
import { first } from 'rxjs/operators'
import { ClientEntry, Controller } from './controller'
import { EMPTY_ENVIRONMENT, Environment } from './environment'

class TestController extends Controller<any, any> {
    public get clientEntries(): Observable<ClientEntry[]> & { value: ClientEntry[] } {
        let value!: ClientEntry[]
        super.clientEntries
            .pipe(first())
            .subscribe(clients => (value = clients))
            .unsubscribe()
        return { ...super.clientEntries, value } as Observable<ClientEntry[]> & { value: ClientEntry[] }
    }
}

const create = (environment?: Environment): TestController => {
    const controller = new TestController({
        clientOptions: () => ({
            createMessageTransports: async () => {
                throw new Error('connection is not used in unit test')
            },
        }),
    })
    if (environment) {
        controller.setEnvironment(environment)
    }
    return controller
}

const FIXTURE_ENVIRONMENT: Environment<any, any> = {
    root: 'file:///',
    component: {
        document: { uri: 'file:///f', languageId: 'l', version: 1, text: '' },
        selections: [],
        visibleRanges: [],
    },
    extensions: [{ id: 'x' }],
    configuration: {},
    context: {},
}

describe('Controller', () => {
    it('creates clients for the environment', () => {
        const controller = create(FIXTURE_ENVIRONMENT)
        assert.deepStrictEqual(
            controller.clientEntries.value.map(({ client }) => ({ id: client.id, root: client.options.root })),
            [{ id: 'x', root: 'file:///' }]
        )
    })

    it('creates clients for extensions even when root and component are not set', () => {
        assert.strictEqual(
            create({ ...EMPTY_ENVIRONMENT, extensions: FIXTURE_ENVIRONMENT.extensions }).clientEntries.value.length,
            1
        )
    })

    it('creates no clients if the environment needs none', () => {
        assert.deepStrictEqual(create(EMPTY_ENVIRONMENT).clientEntries.value, [])
    })
})
