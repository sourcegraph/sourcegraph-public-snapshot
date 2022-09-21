import { act, createEvent, fireEvent, render } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import * as React from 'react'

import { ModifierKey } from '../keys'
import Shortcut from '../Shortcut'
import ShortcutProvider from '../ShortcutProvider'

describe('ShortcutManager', () => {
    let originalGetModifierState = KeyboardEvent.prototype.getModifierState

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

    it('calls the matching shortcut immediately if there are no other similar shortcuts', async () => {
        const fooSpy = jest.fn()
        const barSpy = jest.fn()

        render(
            <ShortcutProvider>
                <Shortcut key="foo" ordered={['f', 'o', 'o']} onMatch={fooSpy} />
                <Shortcut key="bar" ordered={['b', 'a', 'r']} onMatch={barSpy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('foo')

        expect(fooSpy).toHaveBeenCalled()
        expect(barSpy).not.toHaveBeenCalled()
    })

    it('calls multiple shortcuts', () => {
        const fooSpy = jest.fn()
        const barSpy = jest.fn()
        render(
            <ShortcutProvider>
                <Shortcut key="foo" ordered={['f', 'o', 'o']} onMatch={fooSpy} />
                <Shortcut key="bar" ordered={['b', 'a', 'r']} onMatch={barSpy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('foo')
        userEvent.keyboard('bar')

        expect(fooSpy).toHaveBeenCalledTimes(1)
        expect(barSpy).toHaveBeenCalledTimes(1)
    })

    it('matches the longest fully matched shortcut when there are conflicting shortcuts after a timeout', () => {
        const fooSpy = jest.fn()
        const foSpy = jest.fn()
        const fSpy = jest.fn()

        render(
            <ShortcutProvider>
                <Shortcut key="foo" ordered={['f', 'o', 'o']} onMatch={fooSpy} />
                <Shortcut key="fo" ordered={['f', 'o']} onMatch={foSpy} />
                <Shortcut key="f" ordered={['f']} onMatch={fSpy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('fo')

        expect(foSpy).not.toBeCalled()

        jest.runAllTimers()

        expect(fSpy).not.toHaveBeenCalled()
        expect(foSpy).toHaveBeenCalledTimes(1)
        expect(fooSpy).not.toHaveBeenCalled()
    })

    it('does not call shortcuts that do not match the keys pressed', () => {
        const spy = jest.fn()
        render(
            <ShortcutProvider>
                <Shortcut ordered={['b', 'a', 'r']} onMatch={spy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('baz')

        expect(spy).not.toBeCalled()

        jest.runAllTimers()

        expect(spy).not.toBeCalled()
    })

    it('does not call shortcuts that only partially match', () => {
        const spy = jest.fn()
        render(
            <ShortcutProvider>
                <Shortcut key="foo" ordered={['f', 'o', 'o']} onMatch={spy} />
                <Shortcut key="f" ordered={['f']} onMatch={spy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('fo')

        jest.runAllTimers()

        expect(spy).not.toBeCalled()
    })

    it.skip('calls shortcuts that are scoped to a specific node only when that node is focused', () => {
        // This test is meaningless atm because the current implementation of
        // Shortcut doesn't actually work for scoped events.

        const spy = jest.fn()

        act(() => {
            render(
                <ShortcutProvider>
                    <ShortcutWithFocus spy={spy} />
                </ShortcutProvider>
            )
        })

        userEvent.keyboard('z')
        expect(spy).toBeCalled()
    })

    it('only registers a unique shortcut once', () => {
        const spy = jest.fn()

        render(
            <ShortcutProvider>
                <Shortcut key="foo-1" ordered={['f', 'o', 'o']} onMatch={spy} />
                <Shortcut key="foo-2" ordered={['f', 'o', 'o']} onMatch={spy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('foo')

        jest.runAllTimers()

        expect(spy).toHaveBeenCalledTimes(1)
    })

    it('unsubscribes keys when Shortcut unmounts', () => {
        const spy = jest.fn()

        const app = render(
            <ShortcutProvider>
                <Shortcut key="bar" ordered={['b', 'a', 'r']} onMatch={spy} />
                <Shortcut key="foo" ordered={['f', 'o', 'o']} onMatch={spy} />
            </ShortcutProvider>
        )

        app.unmount()

        userEvent.keyboard('foo')
        userEvent.keyboard('bar')

        expect(spy).not.toBeCalled()
    })

    it('resets keys when there are no matching shortcuts', () => {
        const spy = jest.fn()

        render(
            <ShortcutProvider>
                <Shortcut ordered={['?']} onMatch={spy} />
            </ShortcutProvider>
        )

        userEvent.keyboard('{shift}{/shift}a?')

        expect(spy).toHaveBeenCalledTimes(1)
    })

    it('allows default event to occur', () => {
        const spy = jest.fn()

        render(
            <ShortcutProvider>
                <Shortcut ordered={['a']} onMatch={spy} allowDefault />
            </ShortcutProvider>
        )

        const event = createEvent.keyDown(document, { key: 'a' })
        fireEvent(document, event)

        expect(spy).toHaveBeenCalledTimes(1)
        expect(event.defaultPrevented).toBe(false)
    })

    it('prevents the default event by default', () => {
        const spy = jest.fn()

        render(
            <ShortcutProvider>
                <Shortcut ordered={['a']} onMatch={spy} />
            </ShortcutProvider>
        )

        const event = createEvent.keyDown(document, { key: 'a' })
        fireEvent(document, event)

        expect(spy).toHaveBeenCalledTimes(1)
        expect(event.defaultPrevented).toBe(true)
    })

    describe('modifier keys', () => {
        it('matches shortcut when all modifier keys are pressed', () => {
            const fooSpy = jest.fn()
            const held: ModifierKey[] = ['Control', 'Shift', 'Alt', 'Meta']

            render(
                <ShortcutProvider>
                    <Shortcut held={held} ordered={['/']} onMatch={fooSpy} />
                </ShortcutProvider>
            )

            userEvent.keyboard('{Control>}{Shift>}{Alt>}{Meta>}/{/Meta}{/Alt}{/Shift}{/Control}')

            expect(fooSpy).toHaveBeenCalled()
        })

        it('doesn’t match shortcut when all modifier keys not pressed', () => {
            const fooSpy = jest.fn()
            const heldToCheck: ModifierKey[] = ['Control', 'Shift', 'Alt', 'Meta']

            render(
                <ShortcutProvider>
                    <Shortcut held={heldToCheck} ordered={['/']} onMatch={fooSpy} />
                </ShortcutProvider>
            )

            userEvent.keyboard('{Control>}{Shift>}/{/Shift}{/Control}')

            expect(fooSpy).not.toHaveBeenCalled()
        })
    })
})

interface Props {
    spy: jest.Mock<{}>
}

interface State {
    node: HTMLElement | null
}

class ShortcutWithFocus extends React.Component<Props, State> {
    state: State = {
        node: null,
    }

    componentWillUpdate() {
        const { node } = this.state

        if (!node) {
            return
        }

        node.focus()
    }

    render() {
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
