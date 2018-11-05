import { filter, first } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { Controller } from '../client/controller'
import { Environment } from '../client/environment'
import { createExtensionHost } from '../extension/extensionHost'
import { createMessageTransports } from '../protocol/jsonrpc2/helpers.test'

const FIXTURE_ENVIRONMENT: Environment = {
    visibleTextDocuments: [
        {
            uri: 'file:///f',
            languageId: 'l',
            text: 't',
        },
    ],
    extensions: [{ id: 'x' }],
    configuration: { merged: { a: 1 } },
    context: {},
}

interface TestContext {
    clientController: Controller<any, any>
    extensionHost: typeof sourcegraph
}

/**
 * Set up a new client-extension integration test.
 *
 * @internal
 */
export async function integrationTestContext(): Promise<
    TestContext & {
        getEnvironment: () => Environment
        ready: Promise<void>
    }
> {
    const [clientTransports, serverTransports] = createMessageTransports()

    const clientController = new Controller({
        clientOptions: () => ({ createMessageTransports: () => clientTransports }),
    })
    clientController.setEnvironment(FIXTURE_ENVIRONMENT)

    // Ack all configuration updates.
    clientController.configurationUpdates.subscribe(({ resolve }) => resolve(Promise.resolve()))

    const extensionHost = createExtensionHost(
        { bundleURL: '', sourcegraphURL: 'https://example.com', clientApplication: 'sourcegraph' },
        serverTransports
    )

    // Wait for client to be ready.
    await clientController.clientEntries
        .pipe(
            filter(entries => entries.length > 0),
            first()
        )
        .toPromise()

    return {
        clientController,
        extensionHost,
        getEnvironment(): Environment {
            // This runs synchronously because the Observable's root source is a BehaviorSubject (which has an initial value).
            // Confirm it is synchronous just in case, because a bug here would be hard to diagnose.
            let value!: Environment
            let sync = false
            clientController.environment
                .pipe(first())
                .subscribe(environment => {
                    value = environment
                    sync = true
                })
                .unsubscribe()
            if (!sync) {
                throw new Error('environment is not synchronously available')
            }
            return value
        },
        ready: ready({ clientController, extensionHost }),
    }
}

/** @internal */
async function ready({ extensionHost }: TestContext): Promise<void> {
    await extensionHost.internal.sync()
}

/**
 * Returns a {@link Promise} and a function. The {@link Promise} blocks until the returned function is called.
 *
 * @internal
 */
export function createBarrier(): { wait: Promise<void>; done: () => void } {
    let done!: () => void
    const wait = new Promise<void>(resolve => (done = resolve))
    return { wait, done }
}

export function collectSubscribableValues<T>(subscribable: sourcegraph.Subscribable<T>): T[] {
    const values: T[] = []
    subscribable.subscribe(value => values.push(value))
    return values
}
