import React from 'react'
import renderer from 'react-test-renderer'
import { setLinkComponent } from '../../../../../shared/src/components/Link'
import { ListHeaderQueryLinksNav } from './ListHeaderQueryLinks'

// tslint:disable: jsx-no-lambda
describe('ListHeaderQueryLinks', () => {
    setLinkComponent((props: any) => <a {...props} />)
    afterAll(() => setLinkComponent(null as any)) // reset global env for other tests

    test('simple', () =>
        expect(
            renderer
                .create(
                    <ListHeaderQueryLinksNav
                        query="is:b"
                        links={[
                            { label: 'a', queryField: 'is', queryValues: ['a'], count: 1 },
                            { label: 'b', queryField: 'is', queryValues: ['b'], count: 2 },
                        ]}
                        location={{ pathname: '/', hash: '', state: '', search: 'a=b' }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot())
})
