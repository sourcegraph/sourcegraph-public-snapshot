import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { createBarrier } from '../api/integration-test/testHelpers'
import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'
import { ActionItem } from './ActionItem'

describe('ActionItem', () => {
    const NOOP_EXTENSIONS_CONTROLLER = { executeCommand: () => Promise.resolve(undefined) }
    const NOOP_PLATFORM_CONTEXT = { forceUpdateTooltip: () => undefined }
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

    test('pressed toggle actionItem', () => {
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

    test('non-pressed actionItem', () => {
        const component = renderer.create(
            <ActionItem
                action={{ id: 'a', command: 'c', actionItem: { pressed: false, label: 'b' } }}
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
                extensionsController={{ ...NOOP_EXTENSIONS_CONTROLLER, executeCommand: () => wait }}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )

        // Run command and wait for execution to finish.
        let tree = component.toJSON()
        tree!.props.onClick({ preventDefault: () => undefined, currentTarget: { blur: () => undefined } })
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
                extensionsController={{ ...NOOP_EXTENSIONS_CONTROLLER, executeCommand: () => wait }}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )

        // Run command and wait for execution to finish.
        let tree = component.toJSON()
        tree!.props.onClick({ preventDefault: () => undefined, currentTarget: { blur: () => undefined } })
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
                    executeCommand: () => Promise.reject(new Error('x')),
                }}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )

        // Run command (which will reject with an error). (Use setTimeout to wait for the executeCommand resolution
        // to result in the setState call.)
        let tree = component.toJSON()
        tree!.props.onClick({ preventDefault: () => undefined, currentTarget: { blur: () => undefined } })
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
                    executeCommand: () => Promise.reject(new Error('x')),
                }}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )

        // Run command (which will reject with an error). (Use setTimeout to wait for the executeCommand resolution
        // to result in the setState call.)
        let tree = component.toJSON()
        tree!.props.onClick({ preventDefault: () => undefined, currentTarget: { blur: () => undefined } })
        await new Promise<void>(r => setTimeout(r))
        tree = component.toJSON()
        expect(tree).toMatchSnapshot()
    })

    test('render as link for "open" command', () => {
        jsdom.reconfigure({ url: 'https://example.com/foo' })

        const component = renderer.create(
            <ActionItem
                action={{ id: 'c', command: 'open', commandArguments: ['https://example.com/bar'], title: 't' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('render as link with icon for "open" command with different origin', () => {
        jsdom.reconfigure({ url: 'https://example.com/foo' })

        const component = renderer.create(
            <ActionItem
                action={{ id: 'c', command: 'open', commandArguments: ['https://other.com/foo'], title: 't' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
})
