import React from 'react'
import renderer, { act } from 'react-test-renderer'
import sinon from 'sinon'
import { NotificationType } from '../../../../../../shared/src/api/client/services/notifications'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CampaignHeaderEditableName, updateCampaign } from './CampaignHeaderEditableName'

/* eslint-disable @typescript-eslint/no-floating-promises */
describe('CampaignHeaderEditableName', () => {
    const create = ({
        showMessagesNext = sinon.spy(),
        _updateCampaign = sinon.spy(),
    }: {
        showMessagesNext?: () => void
        _updateCampaign?: typeof updateCampaign
    } = {}) =>
        renderer.create(
            <CampaignHeaderEditableName
                campaign={{ id: 'a1', name: 't' }}
                onCampaignUpdate={sinon.spy()}
                className="c"
                extensionsController={{
                    services: {
                        notifications: {
                            showMessages: { next: showMessagesNext },
                        },
                    },
                }}
                _updateCampaign={_updateCampaign}
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

        act(() => e.root.findByProps({ type: 'reset' }).props.onClick({ preventDefault: () => undefined }))
        expect(e.toJSON()).toEqual(viewingSnapshot)
        expect(e.root.findAllByType('input').length).toBe(0)
    })

    test('edit, change, then cancel', () => {
        const e = create()
        const viewingSnapshot = e.toJSON()

        act(() => e.root.findByType('button').props.onClick())
        act(() => {
            e.root.findByType('input').props.onChange({ currentTarget: { value: 't2' } })
            e.root.findByProps({ type: 'reset' }).props.onClick({ preventDefault: () => undefined })
        })
        expect(e.toJSON()).toEqual(viewingSnapshot)
        expect(e.root.findAllByType('input').length).toBe(0)

        act(() => e.root.findByType('button').props.onClick()) // edit again
        expect(e.root.findByType('input').props.value).toBe('t')
    })

    test('save', async () => {
        let resolve: (value: GQL.ICampaign) => void
        const updateCampaignResult = new Promise<GQL.ICampaign>(resolve_ => {
            resolve = resolve_
        })
        const e = create({
            _updateCampaign: ({ input }) => {
                expect(input.id).toBe('a')
                expect(input.name).toBe('t2')
                return updateCampaignResult
            },
        })

        act(() => e.root.findByType('button').props.onClick())
        act(() => e.root.findByType('input').props.onChange({ currentTarget: { value: 't2' } }))
        act(() => {
            e.root.findByType('form').props.onSubmit({ preventDefault: () => undefined })
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
            resolve({ name: 't2' } as any)
        })
        await updateCampaignResult
        expect(e.root.findAllByType('input').length).toBe(0)
    })

    test('error saving', async () => {
        let reject: (error: any) => void
        const updateCampaignResult = new Promise<GQL.ICampaign>((_resolve, reject_) => {
            reject = reject_
        })

        const showMessagesNext = sinon.spy()
        const e = create({ _updateCampaign: () => updateCampaignResult, showMessagesNext })

        act(() => e.root.findByType('button').props.onClick())
        act(() => e.root.findByType('input').props.onChange({ currentTarget: { value: 't2' } }))
        act(() => {
            e.root.findByType('form').props.onSubmit({ preventDefault: () => undefined })
        })

        // TODO: This prints harmless warnings like `Warning: An update to null inside a test was
        // not wrapped in act(...).`. The fix is to use async `act` when it is in the released
        // version of React (see https://github.com/facebook/react/issues/14775).
        act(() => {
            reject({ message: 'x' })
        })
        await updateCampaignResult.catch(() => undefined)
        expect(showMessagesNext.callCount).toBe(1)
        expect(showMessagesNext.firstCall.args).toEqual([
            {
                message: 'Error editing name of campaign: x',
                type: NotificationType.Error,
            },
        ])
        expect(e.root.findAllByType('input').length).toBe(0)
    })
})
