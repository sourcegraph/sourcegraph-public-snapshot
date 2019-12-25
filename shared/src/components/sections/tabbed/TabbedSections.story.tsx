import { storiesOf } from '@storybook/react'
import React, { useEffect } from 'react'
import { Route } from 'react-router'
// eslint-disable-next-line no-restricted-imports
import { BrowserRouter, Link } from 'react-router-dom'
import { setLinkComponent } from '../../Link'
import { Section } from '../Sections'
import './index.scss'
import {
    TabbedSectionsWithLocalStorageViewStatePersistence,
    TabbedSectionsWithURLViewStatePersistence,
} from './TabbedSections'

const { add } = storiesOf('TabbedSections', module).addDecorator(story => (
    // tslint:disable-next-line: jsx-ban-props
    <div style={{ maxWidth: '20rem', margin: '2rem' }}>{story()}</div>
))

type SectionID = 'foo' | 'bar' | 'baz' | 'qux'

const SECTIONS: Section<SectionID>[] = [
    { id: 'foo', label: 'Foo' },
    { id: 'bar', label: 'Bar' },
    {
        id: 'baz',
        label: (
            <div className="flex align-items-center">
                Baz{' '}
                <small style={{ backgroundColor: 'green', color: 'white', borderRadius: '50%', padding: '4px' }}>
                    17
                </small>
            </div>
        ),
    },
    { id: 'qux', label: 'Qux' },
]

add('localStorage', () => (
    <TabbedSectionsWithLocalStorageViewStatePersistence sections={SECTIONS} storageKey="TabbedSections.story.simple">
        <div key="foo">Foo content</div>
        <div key="bar">Bar content</div>
        <div key="baz">Baz content</div>
        <div key="qux">Qux content</div>
    </TabbedSectionsWithLocalStorageViewStatePersistence>
))

add('URL', () => {
    const C: React.FunctionComponent = () => {
        setLinkComponent(Link as any)
        useEffect(
            () => () => {
                setLinkComponent(null)
            },
            []
        )
        return (
            <BrowserRouter>
                <Route
                    render={({ location }) => (
                        <TabbedSectionsWithURLViewStatePersistence sections={SECTIONS} location={location}>
                            <div key="foo">Foo content</div>
                            <div key="bar">Bar content</div>
                            <div key="baz">Baz content</div>
                            <div key="qux">Qux content</div>
                        </TabbedSectionsWithURLViewStatePersistence>
                    )}
                />
            </BrowserRouter>
        )
    }

    return <C />
})
