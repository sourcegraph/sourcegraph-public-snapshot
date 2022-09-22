import * as React from 'react'

import { act, createEvent, fireEvent, render } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import * as sinon from 'sinon'

import { ModifierKey } from './keys'

import { Shortcut, ShortcutProvider } from '.'

describe('ShortcutManager', () => {
    // We only want to preserve the original implementation, not call it as a
    // function.
    // eslint-disable-next-line @typescript-eslint/unbound-method
    const originalGetModifierState = KeyboardEvent.prototype.getModifierState

    beforeAll(() => {
        jest.useFakeTimers()
        // jsdom doesn't implement getModifierState properly:
        // https://github.com/jsdom/jsdom/issues/3126
        KeyboardEvent.prototype.getModifierState = function (key: string): boolean {
            switch (key) {
                case 'Alt':
                    return this.altKey
                case 'Control':
                    return this.ctrlKey
                case 'Meta':
                    return this.metaKey
                case 'Shift':
                    return this.shiftKey
            }
            return false
        }
    })

    afterAll(() => {
        jest.useRealTimers()
        KeyboardEvent.prototype.getModifierState = originalGetModifierState
    })

    it('calls the matching shortcut immediately if there are no other similar shortcuts', () => {
        const fooSpy = sinon.spy()
        const barSpy = sinon.spy()

        render(
            <ShortcutProvider>
                <Shortcut key="foo" ordered={['f', 'o', 'o']} onMatch={fooSpy} />
                <Shortcut key="bar" ordered={['b', 'a', 'r']} onMatch={barSpy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('foo')

        sinon.assert.called(fooSpy)
        sinon.assert.notCalled(barSpy)
    })

    it('calls multiple shortcuts', () => {
        const fooSpy = sinon.spy()
        const barSpy = sinon.spy()

        render(
            <ShortcutProvider>
                <Shortcut key="foo" ordered={['f', 'o', 'o']} onMatch={fooSpy} />
                <Shortcut key="bar" ordered={['b', 'a', 'r']} onMatch={barSpy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('foo')
        userEvent.keyboard('bar')

        sinon.assert.calledOnce(fooSpy)
        sinon.assert.calledOnce(fooSpy)
    })

    it('matches the longest fully matched shortcut when there are conflicting shortcuts after a timeout', () => {
        const fooSpy = sinon.spy()
        const foSpy = sinon.spy()
        const fSpy = sinon.spy()

        render(
            <ShortcutProvider>
                <Shortcut key="foo" ordered={['f', 'o', 'o']} onMatch={fooSpy} />
                <Shortcut key="fo" ordered={['f', 'o']} onMatch={foSpy} />
                <Shortcut key="f" ordered={['f']} onMatch={fSpy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('fo')

        sinon.assert.notCalled(foSpy)

        jest.runAllTimers()

        sinon.assert.notCalled(fSpy)
        sinon.assert.calledOnce(foSpy)
        sinon.assert.notCalled(fooSpy)
    })

    it('does not call shortcuts that do not match the keys pressed', () => {
        const spy = sinon.spy()
        render(
            <ShortcutProvider>
                <Shortcut ordered={['b', 'a', 'r']} onMatch={spy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('baz')

        sinon.assert.notCalled(spy)

        jest.runAllTimers()

        sinon.assert.notCalled(spy)
    })

    it('does not call shortcuts that only partially match', () => {
        const spy = sinon.spy()
        render(
            <ShortcutProvider>
                <Shortcut key="foo" ordered={['f', 'o', 'o']} onMatch={spy} />
                <Shortcut key="f" ordered={['f']} onMatch={spy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('fo')

        jest.runAllTimers()

        sinon.assert.notCalled(spy)
    })

    it.skip('calls shortcuts that are scoped to a specific node only when that node is focused', () => {
        // This test is meaningless atm because the current implementation of
        // Shortcut doesn't actually work for scoped events.

        const spy = sinon.spy()

        act(() => {
            render(
                <ShortcutProvider>
                    <ShortcutWithFocus spy={spy} />
                </ShortcutProvider>
            )
        })

        userEvent.keyboard('z')
        sinon.assert.called(spy)
    })

    it('only registers a unique shortcut once', () => {
        const spy = sinon.spy()

        render(
            <ShortcutProvider>
                <Shortcut key="foo-1" ordered={['f', 'o', 'o']} onMatch={spy} />
                <Shortcut key="foo-2" ordered={['f', 'o', 'o']} onMatch={spy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('foo')

        jest.runAllTimers()

        sinon.assert.calledOnce(spy)
    })

    it('unsubscribes keys when Shortcut unmounts', () => {
        const spy = sinon.spy()

        const app = render(
            <ShortcutProvider>
                <Shortcut key="bar" ordered={['b', 'a', 'r']} onMatch={spy} />
                <Shortcut key="foo" ordered={['f', 'o', 'o']} onMatch={spy} />
            </ShortcutProvider>
        )

        app.unmount()

        userEvent.keyboard('foo')
        userEvent.keyboard('bar')

        sinon.assert.notCalled(spy)
    })

    it('resets keys when there are no matching shortcuts', () => {
        const spy = sinon.spy()

        render(
            <ShortcutProvider>
                <Shortcut ordered={['?']} onMatch={spy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('{shift}{/shift}a?')

        sinon.assert.calledOnce(spy)
    })

    it('allows default event to occur', () => {
        const spy = sinon.spy()

        render(
            <ShortcutProvider>
                <Shortcut ordered={['a']} onMatch={spy} allowDefault={true} />
            </ShortcutProvider>
        )

        const event = createEvent.keyDown(document, { key: 'a' })
        fireEvent(document, event)

        sinon.assert.calledOnce(spy)
        expect(event.defaultPrevented).toBe(false)
    })

    it('prevents the default event by default', () => {
        const spy = sinon.spy()

        render(
            <ShortcutProvider>
                <Shortcut ordered={['a']} onMatch={spy} />
            </ShortcutProvider>
        )

        const event = createEvent.keyDown(document, { key: 'a' })
        fireEvent(document, event)

        sinon.assert.calledOnce(spy)
        expect(event.defaultPrevented).toBe(true)
    })

    describe('modifier keys', () => {
        it('matches shortcut when all modifier keys are pressed', () => {
            const fooSpy = sinon.spy()
            const held: ModifierKey[] = ['Control', 'Shift', 'Alt', 'Meta']

            render(
                <ShortcutProvider>
                    <Shortcut held={held} ordered={['/']} onMatch={fooSpy} />
                </ShortcutProvider>
            )

            userEvent.keyboard('{Control>}{Shift>}{Alt>}{Meta>}/{/Meta}{/Alt}{/Shift}{/Control}')

            sinon.assert.called(fooSpy)
        })

        it('doesn’t match shortcut when all modifier keys not pressed', () => {
            const fooSpy = sinon.spy()
            const heldToCheck: ModifierKey[] = ['Control', 'Shift', 'Alt', 'Meta']

            render(
                <ShortcutProvider>
                    <Shortcut held={heldToCheck} ordered={['/']} onMatch={fooSpy} />
                </ShortcutProvider>
            )

            userEvent.keyboard('{Control>}{Shift>}/{/Shift}{/Control}')

            sinon.assert.notCalled(fooSpy)
        })
    })
})

interface Props {
    spy: sinon.SinonSpy
}

interface State {
    node: HTMLElement | null
}

// This component isn't used currently. It and the corresponding test need
// to be updated.
// eslint-disable-next-line react/no-unsafe
class ShortcutWithFocus extends React.Component<Props, State> {
    public state: State = {
        node: null,
    }

    public UNSAFE_componentWillUpdate(): void {
        const { node } = this.state

        if (!node) {
            return
        }

        node.focus()
    }

    public render(): React.ReactNode {
        const { spy } = this.props
        const { node } = this.state
        return (
            <div className="app">
                <button type="button" ref={this.setRef} />
                <Shortcut ordered={['z']} onMatch={spy} node={node} />
            </div>
        )
    }

    private setRef = (node: HTMLElement | null) => {
        this.setState({
            node,
        })
    }
}
