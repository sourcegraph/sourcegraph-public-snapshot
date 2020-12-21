import { mount } from 'enzyme'
import React, { ChangeEvent } from 'react'
import { Button, Form, Input } from 'reactstrap'
import sinon from 'sinon'
import { Progress } from '../../../stream'
import { StreamingProgressSkippedPopover } from './StreamingProgressSkippedPopover'

describe('StreamingProgressSkippedPopover', () => {
    it('should render correctly', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
                {
                    reason: 'shard-timedout',
                    message: 'Search timed out',
                    severity: 'warn',
                    title: 'Search timed out',
                    suggested: {
                        title: 'timeout:2m',
                        queryExpression: 'timeout:2m',
                    },
                },
            ],
        }

        const element = mount(<StreamingProgressSkippedPopover progress={progress} onSearchAgain={sinon.spy()} />)
        expect(element).toMatchSnapshot()
    })

    it('should not show Search Again section if no suggested searches are set', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                },
            ],
        }

        const element = mount(<StreamingProgressSkippedPopover progress={progress} onSearchAgain={sinon.spy()} />)
        expect(element.find(Form)).toHaveLength(0)
    })

    it('should have Search Again button disabled by default', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
            ],
        }

        const element = mount(<StreamingProgressSkippedPopover progress={progress} onSearchAgain={sinon.spy()} />)
        const searchAgainButton = element.find(Button)
        expect(searchAgainButton).toHaveLength(1)
        expect(searchAgainButton.prop('disabled')).toBe(true)
    })

    it('should enable Search Again button if at least one item is checked', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
                {
                    reason: 'shard-timedout',
                    message: 'Search timed out',
                    severity: 'warn',
                    title: 'Search timed out',
                    suggested: {
                        title: 'timeout:2m',
                        queryExpression: 'timeout:2m',
                    },
                },
            ],
        }

        const element = mount(<StreamingProgressSkippedPopover progress={progress} onSearchAgain={sinon.spy()} />)

        const checkboxes = element.find(Input)
        expect(checkboxes).toHaveLength(3)
        const checkbox = checkboxes.at(1)
        checkbox.invoke('onChange')?.({
            currentTarget: { checked: true, value: checkbox.props().value as string },
        } as ChangeEvent<HTMLInputElement>)

        const searchAgainButton = element.find(Button)
        expect(searchAgainButton).toHaveLength(1)
        expect(searchAgainButton.prop('disabled')).toBe(false)
    })

    it('should disable Search Again button if unchecking all items', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
                {
                    reason: 'shard-timedout',
                    message: 'Search timed out',
                    severity: 'warn',
                    title: 'Search timed out',
                    suggested: {
                        title: 'timeout:2m',
                        queryExpression: 'timeout:2m',
                    },
                },
            ],
        }

        const element = mount(<StreamingProgressSkippedPopover progress={progress} onSearchAgain={sinon.spy()} />)

        const checkboxes = element.find(Input)
        expect(checkboxes).toHaveLength(3)
        const checkbox = checkboxes.at(1)
        checkbox.invoke('onChange')?.({
            currentTarget: { checked: true, value: checkbox.props().value as string },
        } as ChangeEvent<HTMLInputElement>)

        let searchAgainButton = element.find(Button)
        expect(searchAgainButton.prop('disabled')).toBe(false)

        checkbox.invoke('onChange')?.({
            currentTarget: { checked: false, value: checkbox.props().value as string },
        } as ChangeEvent<HTMLInputElement>)

        searchAgainButton = element.find(Button)
        expect(searchAgainButton.prop('disabled')).toBe(true)
    })

    it('should call onSearchAgain with selected items when button is clicked', () => {
        const progress: Progress = {
            durationMs: 1500,
            matchCount: 2,
            repositoriesCount: 2,
            skipped: [
                {
                    reason: 'excluded-fork',
                    message: '10k forked repositories excluded',
                    severity: 'info',
                    title: '10k forked repositories excluded',
                    suggested: {
                        title: 'forked:yes',
                        queryExpression: 'forked:yes',
                    },
                },
                {
                    reason: 'excluded-archive',
                    message: '60k archived repositories excluded',
                    severity: 'info',
                    title: '60k archived repositories excluded',
                    suggested: {
                        title: 'archived:yes',
                        queryExpression: 'archived:yes',
                    },
                },
                {
                    reason: 'shard-timedout',
                    message: 'Search timed out',
                    severity: 'warn',
                    title: 'Search timed out',
                    suggested: {
                        title: 'timeout:2m',
                        queryExpression: 'timeout:2m',
                    },
                },
            ],
        }

        const searchAgain = sinon.spy()

        const element = mount(<StreamingProgressSkippedPopover progress={progress} onSearchAgain={searchAgain} />)

        const checkboxes = element.find(Input)

        expect(checkboxes).toHaveLength(3)
        const checkbox1 = checkboxes.at(1)
        checkbox1.invoke('onChange')?.({
            currentTarget: { checked: true, value: checkbox1.props().value as string },
        } as ChangeEvent<HTMLInputElement>)

        expect(checkboxes).toHaveLength(3)
        const checkbox2 = checkboxes.at(2)
        checkbox2.invoke('onChange')?.({
            currentTarget: { checked: true, value: checkbox2.props().value as string },
        } as ChangeEvent<HTMLInputElement>)

        const form = element.find(Form)
        form.simulate('submit')

        sinon.assert.calledOnce(searchAgain)
        sinon.assert.calledWith(searchAgain, ['archived:yes', 'timeout:2m'])
    })
})
