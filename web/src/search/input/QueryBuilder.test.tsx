import * as React from 'react'
import { cleanup, fireEvent, getBySelectText, getByTestId, queryByTestId, render, wait } from 'react-testing-library'
import sinon from 'sinon'
import { QueryBuilder } from './QueryBuilder'

describe('QueryBuilder', () => {
    afterAll(cleanup)

    it('fires the onFieldsQueryChange prop handler with the `repo:` filter when updating the "Repository" field', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        const repoField = container.querySelector('#query-builder-repo')!
        fireEvent.change(repoField, { target: { value: 'sourcegraph/sourcegraph' } })
        expect(onChange.calledOnce).toBe(true)
        expect(onChange.calledWith('repo:sourcegraph/sourcegraph')).toBe(true)
    })

    it('fires the onFieldsQueryChange prop handler with the `file:` filter when updating the "File" field', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        const fileField = container.querySelector('#query-builder-file')!
        fireEvent.change(fileField, { target: { value: 'web' } })
        expect(onChange.calledOnce).toBe(true)
        expect(onChange.calledWith('file:web')).toBe(true)
    })

    it('fires the onFieldsQueryChange prop handler with the patterns left untransformed when updating the "Patterns" field', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        const patternsField = container.querySelector('#query-builder-patterns')!
        fireEvent.change(patternsField, { target: { value: '(open|close) file' } })
        expect(onChange.calledOnce).toBe(true)
        expect(onChange.calledWith('(open|close) file')).toBe(true)
    })

    it('field fires the onFieldsQueryChange prop handler with the term wrapped in double quotes when updating the "Exact match"', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        const exactMatchField = container.querySelector('#query-builder-exactMatch')!
        fireEvent.change(exactMatchField, { target: { value: 'foo bar baz' } })
        expect(onChange.calledOnce).toBe(true)
        expect(onChange.calledWith('"foo bar baz"')).toBe(true)
    })

    it('checks that the "Author", "Before", "After", and "Message" fields do not exist if the search type is set to text search', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        expect(queryByTestId(container, 'test-author')).toBeNull()
        expect(queryByTestId(container, 'test-after')).toBeNull()
        expect(queryByTestId(container, 'test-before')).toBeNull()
        expect(queryByTestId(container, 'test-message')).toBeNull()
    })

    it('checks that the "Author", "Before", "After", and "Message" fields exist if the search type is set to diff search', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'diff' } })

        await wait(() => queryByTestId(container, 'test-author'))
        expect(queryByTestId(container, 'test-author')).toBeTruthy()
        await wait(() => queryByTestId(container, 'test-after'))
        expect(queryByTestId(container, 'test-after')).toBeTruthy()
        await wait(() => queryByTestId(container, 'test-before'))
        expect(queryByTestId(container, 'test-before')).toBeTruthy()
        await wait(() => queryByTestId(container, 'test-message'))
        expect(queryByTestId(container, 'test-message')).toBeTruthy()
    })

    it('checks that the "Message" field does not exist if type is commit', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'commit' } })

        expect(queryByTestId(container, 'test-message')).toBeNull()
    })

    it('checks that the "Author", "Before", and "After" fields exist if type is commit', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'commit' } })

        await wait(() => queryByTestId(container, 'test-author'))
        expect(queryByTestId(container, 'test-author')).toBeTruthy()
        await wait(() => queryByTestId(container, 'test-after'))
        expect(queryByTestId(container, 'test-after')).toBeTruthy()
        await wait(() => queryByTestId(container, 'test-before'))
        expect(queryByTestId(container, 'test-before')).toBeTruthy()
        await wait(() => queryByTestId(container, 'test-message'))
    })

    it('fires the onFieldsQueryChange prop handler with the "author:" filter when updating the "Author" field', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'diff' } })

        await wait(() => queryByTestId(container, 'test-author'))
        const authorField = container.querySelector('#query-builder-author')!
        fireEvent.change(authorField, { target: { value: 'alice' } })
        expect(onChange.calledTwice).toBe(true)
        expect(onChange.calledWith('type:diff author:alice')).toBe(true)
    })

    it('fires the onFieldsQueryChange prop handler with the "after:" filter when updating the "After" field ', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'diff' } })

        await wait(() => queryByTestId(container, 'test-after'))
        const afterField = container.querySelector('#query-builder-after')!
        fireEvent.change(afterField, { target: { value: '1 year ago' } })

        expect(onChange.calledTwice).toBe(true)
        expect(onChange.calledWith('type:diff after:"1 year ago"')).toBe(true)
    })

    it('fires the onFieldsQueryChange prop handler with the "before:" filter when updating the "Before" field', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'diff' } })

        await wait(() => queryByTestId(container, 'test-before'))
        const beforeField = container.querySelector('#query-builder-before')!
        fireEvent.change(beforeField, { target: { value: '1 year ago' } })

        expect(onChange.calledTwice).toBe(true)
        expect(onChange.calledWith('type:diff before:"1 year ago"')).toBe(true)
    })

    it('fires the onFieldsQueryChange prop handler with the "message:" filter when updating the "Message" field', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const toggle = getByTestId(container, 'test-query-builder-toggle')
        fireEvent.click(toggle)

        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'diff' } })

        await wait(() => queryByTestId(container, 'test-message'))
        const messageField = container.querySelector('#query-builder-message')!
        fireEvent.change(messageField, { target: { value: 'fix issue' } })

        expect(onChange.calledTwice).toBe(true)
        expect(onChange.calledWith('type:diff message:"fix issue"')).toBe(true)
    })
})
