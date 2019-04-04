import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { createBarrier } from '../api/integration-test/testHelpers'
import { setLinkComponent } from '../components/Link'
import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'
import { ActionItem } from './ActionItem'

describe('ActionItem', () => {
    const NOOP_EXTENSIONS_CONTROLLER = { executeCommand: async () => void 0 }
    const NOOP_PLATFORM_CONTEXT = { forceUpdateTooltip: () => void 0 }
    const history = H.createMemoryHistory()

    test('non-actionItem variant', () => {
        const component = renderer.create(
            <ActionItem
                action={{ id: 'c', command: 'c', title: 't', description: 'd', iconURL: 'u', category: 'g' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('actionItem variant', () => {
        const component = renderer.create(
            <ActionItem
                action={{ id: 'c', command: 'c', title: 't', description: 'd', iconURL: 'u', category: 'g' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                variant="actionItem"
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('noop command', () => {
        const component = renderer.create(
            <ActionItem
                action={{ id: 'c', title: 't', description: 'd', iconURL: 'u', category: 'g' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('pressed actionItem', () => {
        const component = renderer.create(
            <ActionItem
                action={{ id: 'a', command: 'c', actionItem: { pressed: true, label: 'b' } }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                variant="actionItem"
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('title element', () => {
        const component = renderer.create(
            <ActionItem
                action={{ id: 'c', command: 'c', title: 't', description: 'd', iconURL: 'u', category: 'g' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                variant="actionItem"
                title={<span>t2</span>}
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('run command', async () => {
        const { wait, done } = createBarrier()

        const component = renderer.create(
            <ActionItem
                action={{ id: 'c', command: 'c', title: 't', description: 'd', iconURL: 'u', category: 'g' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                variant="actionItem"
                disabledDuringExecution={true}
                location={history.location}
                extensionsController={{ ...NOOP_EXTENSIONS_CONTROLLER, executeCommand: async () => wait }}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )

        // Run command and wait for execution to finish.
        let tree = component.toJSON()
        tree!.props.onClick({ preventDefault: () => void 0, currentTarget: { blur: () => void 0 } })
        tree = component.toJSON()
        expect(tree).toMatchSnapshot()

        // Finish execution. (Use setTimeout to wait for the executeCommand resolution to result in the setState
        // call.)
        done()
        await new Promise<void>(r => setTimeout(r))
        tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })

    test('run command with showLoadingSpinnerDuringExecution', async () => {
        const { wait, done } = createBarrier()

        const component = renderer.create(
            <ActionItem
                action={{ id: 'c', command: 'c', title: 't', description: 'd', iconURL: 'u', category: 'g' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                variant="actionItem"
                showLoadingSpinnerDuringExecution={true}
                location={history.location}
                extensionsController={{ ...NOOP_EXTENSIONS_CONTROLLER, executeCommand: async () => wait }}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )

        // Run command and wait for execution to finish.
        let tree = component.toJSON()
        tree!.props.onClick({ preventDefault: () => void 0, currentTarget: { blur: () => void 0 } })
        tree = component.toJSON()
        expect(tree).toMatchSnapshot()

        // Finish execution. (Use setTimeout to wait for the executeCommand resolution to result in the setState
        // call.)
        done()
        await new Promise<void>(r => setTimeout(r))
        tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })

    test('run command with error', async () => {
        const component = renderer.create(
            <ActionItem
                action={{ id: 'c', command: 'c', title: 't', description: 'd', iconURL: 'u', category: 'g' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                variant="actionItem"
                disabledDuringExecution={true}
                location={history.location}
                extensionsController={{
                    ...NOOP_EXTENSIONS_CONTROLLER,
                    executeCommand: async () => Promise.reject('x'),
                }}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )

        // Run command (which will reject with an error). (Use setTimeout to wait for the executeCommand resolution
        // to result in the setState call.)
        let tree = component.toJSON()
        tree!.props.onClick({ preventDefault: () => void 0, currentTarget: { blur: () => void 0 } })
        await new Promise<void>(r => setTimeout(r))
        tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })

    test('run command with error with showInlineError', async () => {
        const component = renderer.create(
            <ActionItem
                action={{ id: 'c', command: 'c', title: 't', description: 'd', iconURL: 'u', category: 'g' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                variant="actionItem"
                showInlineError={true}
                location={history.location}
                extensionsController={{
                    ...NOOP_EXTENSIONS_CONTROLLER,
                    executeCommand: async () => Promise.reject('x'),
                }}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )

        // Run command (which will reject with an error). (Use setTimeout to wait for the executeCommand resolution
        // to result in the setState call.)
        let tree = component.toJSON()
        tree!.props.onClick({ preventDefault: () => void 0, currentTarget: { blur: () => void 0 } })
        await new Promise<void>(r => setTimeout(r))
        tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })

    test('render as link for "open" command', () => {
        setLinkComponent((props: any) => <a {...props} />)
        afterAll(() => setLinkComponent(null as any)) // reset global env for other tests

        const component = renderer.create(
            <ActionItem
                action={{ id: 'c', command: 'open', commandArguments: ['https://example.com'], title: 't' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
})
