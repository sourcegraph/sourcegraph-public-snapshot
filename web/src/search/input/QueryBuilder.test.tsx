import * as React from 'react'
import { cleanup, fireEvent, getBySelectText, queryByTestId, render, wait } from 'react-testing-library'
import sinon from 'sinon'
import { QueryBuilder } from './QueryBuilder'

describe.only('QueryBuilder', () => {
    afterAll(cleanup)

    it('updating the "Repository" field fires the onFieldsQueryChange prop handler with the `repo:` filter', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)

        const repoField = container.querySelector('#query-builder__repo')!
        fireEvent.change(repoField, { target: { value: 'sourcegraph/sourcegraph' } })
        expect(onChange.calledOnce).toBe(true)
        expect(onChange.calledWith('repo:sourcegraph/sourcegraph')).toBe(true)
    })

    it('updating the "File" field fires the onFieldsQueryChange prop handler with the `file:` filter', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)

        const fileField = container.querySelector('#query-builder__file')!
        fireEvent.change(fileField, { target: { value: 'web' } })
        expect(onChange.calledOnce).toBe(true)
        expect(onChange.calledWith('file:web')).toBe(true)
    })

    it('updating the "Patterns" field fires the onFieldsQueryChange prop handler with the patterns left untransformed', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)

        const patternsField = container.querySelector('#query-builder__patterns')!
        fireEvent.change(patternsField, { target: { value: '(open|close) file' } })
        expect(onChange.calledOnce).toBe(true)
        expect(onChange.calledWith('(open|close) file')).toBe(true)
    })

    it('updating the "Exact match" field fires the onFieldsQueryChange prop handler with the term wrapped in double quotes', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)

        const exactMatchField = container.querySelector('#query-builder__quoted-term')!
        fireEvent.change(exactMatchField, { target: { value: 'foo bar baz' } })
        expect(onChange.calledOnce).toBe(true)
        expect(onChange.calledWith('"foo bar baz"')).toBe(true)
    })

    it('the "Author", "Before", "After", and "Message" fields do not exist if the search type is set to text search', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)

        expect(queryByTestId(container, 'test-author')).toBeNull()
        expect(queryByTestId(container, 'test-after')).toBeNull()
        expect(queryByTestId(container, 'test-before')).toBeNull()
        expect(queryByTestId(container, 'test-message')).toBeNull()
    })

    it('the "Author", "Before", "After", and "Message" fields exist if the search type is set to diff search', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
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

    it('the "Message" field does not exist if type is commit', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'commit' } })

        expect(queryByTestId(container, 'test-message')).toBeNull()
    })

    it('the "Author", "Before", and "After" fields exist if type is commit', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
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

    it('updating the "Author" field fires the onFieldsQueryChange prop handler with the "author:" filter', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'diff' } })

        await wait(() => queryByTestId(container, 'test-author'))
        const authorField = container.querySelector('#query-builder__author')!
        fireEvent.change(authorField, { target: { value: 'alice' } })
        expect(onChange.calledTwice).toBe(true)
        expect(onChange.calledWith('type:diff author:alice')).toBe(true)
    })

    it('updating the "After" field fires the onFieldsQueryChange prop handler with the "after:" filter', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'diff' } })

        await wait(() => queryByTestId(container, 'test-after'))
        const afterField = container.querySelector('#query-builder__after')!
        fireEvent.change(afterField, { target: { value: '1 year ago' } })

        expect(onChange.calledTwice).toBe(true)
        expect(onChange.calledWith('type:diff after:"1 year ago"')).toBe(true)
    })

    it('updating the "Before" field fires the onFieldsQueryChange prop handler with the "before:" filter', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'diff' } })

        await wait(() => queryByTestId(container, 'test-before'))
        const beforeField = container.querySelector('#query-builder__before')!
        fireEvent.change(beforeField, { target: { value: '1 year ago' } })

        expect(onChange.calledTwice).toBe(true)
        expect(onChange.calledWith('type:diff before:"1 year ago"')).toBe(true)
    })

    it('updating the "Message" field fires the onFieldsQueryChange prop handler with the "message:" filter', async () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} isSourcegraphDotCom={false} />)
        const typeField = getBySelectText(container, 'Text (default)')!
        fireEvent.change(typeField, { target: { value: 'diff' } })

        await wait(() => queryByTestId(container, 'test-message'))
        const messageField = container.querySelector('#query-builder__message')!
        fireEvent.change(messageField, { target: { value: 'fix issue' } })

        expect(onChange.calledTwice).toBe(true)
        expect(onChange.calledWith('type:diff message:"fix issue"')).toBe(true)
    })
})
