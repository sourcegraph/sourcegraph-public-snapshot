import React from 'react'
import renderer, { act } from 'react-test-renderer'
import { ViewPage } from './ViewPage'
import * as H from 'history'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import { Controller } from '../../../shared/src/extensions/controller'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { of } from 'rxjs'

jest.mock('@sourcegraph/react-loading-spinner', () => ({ LoadingSpinner: 'LoadingSpinner' }))
jest.mock('./QueryInputInViewContent', () => ({ QueryInputInViewContent: 'QueryInputInViewContent' }))

const commonProps: Omit<React.ComponentProps<typeof ViewPage>, 'viewID' | 'extraPath' | '_getView'> = {
    settingsCascade: { final: {}, subjects: [] },
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    setCaseSensitivity: () => undefined,
    setPatternType: () => undefined,
    history: H.createMemoryHistory(),
    location: H.createLocation('/'),
    extensionsController: { services: { contribution: { getContributions: () => ({}) } } } as Controller,
}

describe('ViewPage', () => {
    test('view is loading', () => {
        expect(
            renderer
                .create(<ViewPage {...commonProps} viewID="v" extraPath="" _getView={() => of(undefined)} />)
                .toJSON()
        ).toMatchSnapshot()
    })

    test('view not found', () => {
        const component = renderer.create(
            <ViewPage {...commonProps} viewID="v" extraPath="" _getView={() => of(null)} />
        )
        act(() => undefined)
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('renders view', () => {
        const component = renderer.create(
            <ViewPage
                {...commonProps}
                viewID="v"
                extraPath=""
                _getView={() =>
                    of({
                        title: 't',
                        content: [
                            {
                                kind: MarkupKind.Markdown,
                                value: '**a**',
                            },
                            {
                                kind: MarkupKind.PlainText,
                                value: '*b*',
                            },
                            {
                                value: '*c*',
                            },
                            {
                                component: 'QueryInput',
                                props: { implicitQueryPrefix: 'x' },
                            },
                        ],
                    })
                }
            />
        )
        act(() => undefined)
        expect(component.toJSON()).toMatchSnapshot()
    })
})
