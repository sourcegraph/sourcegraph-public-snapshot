import { assert, expect } from 'chai'
import { describe, it } from 'mocha'
import * as React from 'react'
import { cleanup, fireEvent, render } from 'react-testing-library'
import { EMPTY, merge, noop, of, Subject } from 'rxjs'
import { switchMap, tap } from 'rxjs/operators'
import { TestScheduler } from 'rxjs/testing'
import sinon from 'sinon'

import { ServerURLForm, ServerURLFormProps } from './ServerURLForm'

describe('ServerURLForm', () => {
    after(cleanup)

    it('fires the onChange prop handler', () => {
        const onChange = sinon.spy()
        const onSubmit = sinon.spy()

        const { container } = render(
            <ServerURLForm
                value={'https://sourcegraph.com'}
                status={'connected'}
                onChange={onChange}
                onSubmit={onSubmit}
            />
        )

        const urlInput = container.querySelector('input')!

        fireEvent.change(urlInput, { target: { value: 'https://different.com' } })

        assert.isTrue(onChange.calledOnce)
        assert.isTrue(onChange.calledWith('https://different.com'))

        assert.isTrue(onSubmit.notCalled)
    })

    it('updates the input value when the url changes', () => {
        const props: ServerURLFormProps = {
            value: 'https://sourcegraph.com',
            status: 'connected',
            onChange: noop,
            onSubmit: noop,
        }

        const { container, rerender } = render(<ServerURLForm {...props} />)

        const urlInput = container.querySelector('input')!

        rerender(<ServerURLForm {...props} value={'https://different.com'} />)

        const newValue = urlInput.value

        expect(newValue).to.equal('https://different.com')
    })

    it('fires the onSubmit prop handler when the form is submitted', () => {
        const onSubmit = sinon.spy()

        const { container } = render(
            <ServerURLForm value={'https://sourcegraph.com'} status={'connected'} onChange={noop} onSubmit={onSubmit} />
        )

        const form = container.querySelector('form')!

        fireEvent.submit(form)

        assert.isTrue(onSubmit.calledOnce)
    })

    it('fires the onSubmit prop handler after 5s on inactivity after a change', () => {
        const scheduler = new TestScheduler((a, b) => assert.deepEqual(a, b))

        scheduler.run(({ cold, expectObservable }) => {
            const submits = new Subject<void>()
            const nextSubmit = () => submits.next()

            const { container } = render(
                <ServerURLForm
                    value={'https://sourcegraph.com'}
                    status={'connected'}
                    onChange={noop}
                    onSubmit={nextSubmit}
                />
            )

            const form = container.querySelector('input')!

            const urls: { [key: string]: string } = {
                a: 'https://different.com',
            }

            const submitObs = cold('a', urls).pipe(
                switchMap(url => {
                    const emit = of(undefined).pipe(
                        tap(() => {
                            fireEvent.change(form, { target: { value: url } })
                        }),
                        switchMap(() => EMPTY)
                    )

                    return merge(submits, emit)
                })
            )

            expectObservable(submitObs).toBe('5s a', { a: undefined })
        })
    })

    it("doesn't submit after 5 seconds if the form was submitted manually", () => {
        const scheduler = new TestScheduler((a, b) => assert.deepEqual(a, b))

        scheduler.run(({ cold, expectObservable }) => {
            const changes = new Subject<string>()
            const nextChange = () => changes.next()

            const submits = new Subject<void>()
            const nextSubmit = () => submits.next()

            const props: ServerURLFormProps = {
                value: 'https://sourcegraph.com',
                status: 'connected',
                onChange: nextChange,
                onSubmit: nextSubmit,
            }

            const { container } = render(<ServerURLForm {...props} />)
            const form = container.querySelector('input')!

            changes.subscribe(url => {
                fireEvent.submit(form)
            })

            const urls: { [key: string]: string } = {
                a: 'https://different.com',
            }

            const submitObs = cold('a', urls).pipe(
                switchMap(url => {
                    const emit = of(undefined).pipe(
                        tap(() => {
                            fireEvent.change(form, { target: { value: url } })
                        }),
                        switchMap(() => EMPTY)
                    )

                    return merge(submits, emit)
                })
            )

            expectObservable(submitObs).toBe('a', { a: undefined })
        })
    })
})
