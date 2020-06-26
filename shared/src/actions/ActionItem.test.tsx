import * as H from 'history'
import React from 'react'
import { createBarrier } from '../api/integration-test/testHelpers'
import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'
import { ActionItem } from './ActionItem'
import { NEVER } from 'rxjs'
import { mount } from 'enzyme'

jest.mock('mdi-react/OpenInNewIcon', () => 'OpenInNewIcon')

describe('ActionItem', () => {
    const NOOP_EXTENSIONS_CONTROLLER = { executeCommand: () => Promise.resolve(undefined) }
    const NOOP_PLATFORM_CONTEXT = { forceUpdateTooltip: () => undefined, settings: NEVER }
    const history = H.createMemoryHistory()

    test('non-actionItem variant', () => {
        const component = mount(
            <ActionItem
                action={{ id: 'c', command: 'c', title: 't', description: 'd', iconURL: 'u', category: 'g' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.children()).toMatchSnapshot()
    })

    test('actionItem variant', () => {
        const component = mount(
            <ActionItem
                action={{ id: 'c', command: 'c', title: 't', description: 'd', iconURL: 'u', category: 'g' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                variant="actionItem"
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.children()).toMatchSnapshot()
    })

    test('noop command', () => {
        const component = mount(
            <ActionItem
                action={{ id: 'c', title: 't', description: 'd', iconURL: 'u', category: 'g' }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.children()).toMatchSnapshot()
    })

    test('pressed toggle actionItem', () => {
        const component = mount(
            <ActionItem
                action={{ id: 'a', command: 'c', actionItem: { pressed: true, label: 'b' } }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                variant="actionItem"
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.children()).toMatchSnapshot()
    })

    test('non-pressed actionItem', () => {
        const component = mount(
            <ActionItem
                action={{ id: 'a', command: 'c', actionItem: { pressed: false, label: 'b' } }}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                variant="actionItem"
                location={history.location}
                extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )
        expect(component.children()).toMatchSnapshot()
    })

    test('title element', () => {
        const component = mount(
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
        expect(component.children()).toMatchSnapshot()
    })

    test('run command', async () => {
        const { wait, done } = createBarrier()

        const component = mount(
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
        component.simulate('click')
        expect(component.children()).toMatchSnapshot()

        // Finish execution. (Use setTimeout to wait for the executeCommand resolution to result in the setState
        // call.)
        done()
        await new Promise<void>(resolve => setTimeout(resolve))
        expect(component.children()).toMatchSnapshot()
    })

    test('run command with showLoadingSpinnerDuringExecution', async () => {
        const { wait, done } = createBarrier()

        const component = mount(
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

        component.simulate('click')
        expect(component.children()).toMatchSnapshot()

        // Finish execution. (Use setTimeout to wait for the executeCommand resolution to result in the setState
        // call.)
        done()
        await new Promise<void>(resolve => setTimeout(resolve))
        expect(component.children()).toMatchSnapshot()
    })

    test('run command with error', async () => {
        const component = mount(
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
        component.simulate('click')
        await new Promise<void>(resolve => setTimeout(resolve))
        expect(component.children()).toMatchSnapshot()
    })

    test('run command with error with showInlineError', async () => {
        const component = mount(
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
        component.simulate('click')
        await new Promise<void>(resolve => setTimeout(resolve))
        expect(component.children()).toMatchSnapshot()
    })

    describe('"open" command', () => {
        it('renders as link', () => {
            jsdom.reconfigure({ url: 'https://example.com/foo' })

            const component = mount(
                <ActionItem
                    action={{ id: 'c', command: 'open', commandArguments: ['https://example.com/bar'], title: 't' }}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    location={history.location}
                    extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                />
            )
            expect(component.children()).toMatchSnapshot()
        })

        it('renders as link with icon and opens a new tab for a different origin', () => {
            jsdom.reconfigure({ url: 'https://example.com/foo' })

            const component = mount(
                <ActionItem
                    action={{ id: 'c', command: 'open', commandArguments: ['https://other.com/foo'], title: 't' }}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    location={history.location}
                    extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                />
            )
            expect(component.children()).toMatchSnapshot()
        })

        it('renders as link that opens in a new tab, but without icon for a different origin as the alt action and a primary action defined', () => {
            jsdom.reconfigure({ url: 'https://example.com/foo' })

            const component = mount(
                <ActionItem
                    action={{ id: 'c1', command: 'whatever', title: 'primary' }}
                    altAction={{ id: 'c2', command: 'open', commandArguments: ['https://other.com/foo'], title: 'alt' }}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    location={history.location}
                    extensionsController={NOOP_EXTENSIONS_CONTROLLER}
                    platformContext={NOOP_PLATFORM_CONTEXT}
                />
            )
            expect(component.children()).toMatchSnapshot()
        })
    })
})
