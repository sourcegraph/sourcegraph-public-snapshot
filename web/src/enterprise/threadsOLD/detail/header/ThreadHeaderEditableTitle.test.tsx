import React from 'react'
import renderer, { act } from 'react-test-renderer'
import sinon from 'sinon'
import { NotificationType } from '../../../../../../shared/src/api/client/services/notifications'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { updateThread } from '../../../../discussions/backend'
import { ThreadHeaderEditableTitle } from './ThreadHeaderEditableTitle'

describe('ThreadHeaderEditableTitle', () => {
    const create = ({
        showMessagesNext = sinon.spy(),
        _updateThread = sinon.spy(),
    }: {
        showMessagesNext?: () => void
        _updateThread?: typeof updateThread
    } = {}) =>
        renderer.create(
            <ThreadHeaderEditableTitle
                thread={{ id: 'a1', idWithoutKind: '1', title: 't' }}
                onThreadUpdate={sinon.spy()}
                className="c"
                extensionsController={{
                    services: {
                        notifications: {
                            showMessages: { next: showMessagesNext },
                        },
                    },
                }}
                _updateThread={_updateThread}
            />
        )

    test('viewing state', () => expect(create().toJSON()).toMatchSnapshot())

    test('enter editing state', () => {
        const e = create()
        act(() => e.root.findByType('button').props.onClick())
        expect(e.toJSON()).toMatchSnapshot()
        expect(e.root.findByType('input')).toBeTruthy()
        expect(e.root.findByType('input').props.value).toBe('t')
    })

    test('edit then immediately cancel', () => {
        const e = create()
        const viewingSnapshot = e.toJSON()

        act(() => e.root.findByType('button').props.onClick())
        expect(e.root.findByType('input')).toBeTruthy()

        act(() => e.root.findByProps({ type: 'reset' }).props.onClick({ preventDefault: () => void 0 }))
        expect(e.toJSON()).toEqual(viewingSnapshot)
        expect(e.root.findAllByType('input').length).toBe(0)
    })

    test('edit, change, then cancel', () => {
        const e = create()
        const viewingSnapshot = e.toJSON()

        act(() => e.root.findByType('button').props.onClick())
        act(() => {
            e.root.findByType('input').props.onChange({ currentTarget: { value: 't2' } })
            e.root.findByProps({ type: 'reset' }).props.onClick({ preventDefault: () => void 0 })
        })
        expect(e.toJSON()).toEqual(viewingSnapshot)
        expect(e.root.findAllByType('input').length).toBe(0)

        act(() => e.root.findByType('button').props.onClick()) // edit again
        expect(e.root.findByType('input').props.value).toBe('t')
    })

    test('save', async () => {
        let resolve: (value: GQL.IDiscussionThread) => void
        const updateThreadResult = new Promise<GQL.IDiscussionThread>(resolve_ => {
            resolve = resolve_
        })
        const e = create({
            _updateThread: input => {
                expect(input.threadID).toBe('a')
                expect(input.title).toBe('t2')
                return updateThreadResult
            },
        })

        act(() => e.root.findByType('button').props.onClick())
        act(() => e.root.findByType('input').props.onChange({ currentTarget: { value: 't2' } }))
        act(() => {
            e.root.findByType('form').props.onSubmit({ preventDefault: () => void 0 })
        })

        // Loading.
        expect(e.root.findByType('input').props.disabled).toBeTruthy()
        expect(e.root.findByProps({ type: 'submit' }).props.disabled).toBeTruthy()
        expect(e.root.findByProps({ type: 'reset' }).props.disabled).toBeTruthy()

        // Successfully edited.
        //
        // TODO: This prints harmless warnings like `Warning: An update to null inside a test was
        // not wrapped in act(...).`. The fix is to use async `act` when it is in the released
        // version of React (see https://github.com/facebook/react/issues/14775).
        act(() => {
            resolve({ title: 't2' } as any)
        })
        await updateThreadResult
        expect(e.root.findAllByType('input').length).toBe(0)
    })

    test('error saving', async () => {
        let reject: (error: any) => void
        const updateThreadResult = new Promise<GQL.IDiscussionThread>((_resolve, reject_) => {
            reject = reject_
        })

        const showMessagesNext = sinon.spy()
        const e = create({ _updateThread: () => updateThreadResult, showMessagesNext })

        act(() => e.root.findByType('button').props.onClick())
        act(() => e.root.findByType('input').props.onChange({ currentTarget: { value: 't2' } }))
        act(() => {
            e.root.findByType('form').props.onSubmit({ preventDefault: () => void 0 })
        })

        // TODO: This prints harmless warnings like `Warning: An update to null inside a test was
        // not wrapped in act(...).`. The fix is to use async `act` when it is in the released
        // version of React (see https://github.com/facebook/react/issues/14775).
        act(() => {
            reject({ message: 'x' })
        })
        await updateThreadResult.catch(() => void 0)
        expect(showMessagesNext.callCount).toBe(1)
        expect(showMessagesNext.firstCall.args).toEqual([
            {
                message: 'Error editing title of thread: x',
                type: NotificationType.Error,
            },
        ])
        expect(e.root.findAllByType('input').length).toBe(0)
    })
})
