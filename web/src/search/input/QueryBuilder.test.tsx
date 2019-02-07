import * as React from 'react'
import { cleanup, fireEvent, render } from 'react-testing-library'
import sinon from 'sinon'
import { QueryBuilder } from './QueryBuilder'

describe.only('QueryBuilder', () => {
    afterAll(cleanup)

    it('updating repo field fires the onQueryChange prop handler', () => {
        const onChange = sinon.spy()
        const { container } = render(<QueryBuilder onFieldsQueryChange={onChange} />)

        const repoField = container.querySelector('#query-builder__repo')!
        fireEvent.change(repoField, { target: { value: 'sourcegraph/sourcegraph' } })
        expect(onChange.calledOnce).toBe(true)
        expect(onChange.calledWith('repo:sourcegraph/sourcegraph')).toBe(true)
    })
})
