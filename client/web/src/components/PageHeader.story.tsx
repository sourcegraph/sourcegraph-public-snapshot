import { storiesOf } from '@storybook/react'
import React from 'react'
import SearchIcon from 'mdi-react/SearchIcon'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'

import webStyles from '../SourcegraphWebApp.scss'
import { PageHeader } from './PageHeader'
import { Link } from '../../../shared/src/components/Link'
import { WebStory } from './WebStory'

const { add } = storiesOf('web/PageHeader', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container web-content mt-3">{story()}</div>
        </div>
    </>
))

add('Basic header', () => (
    <WebStory>
        {() => (
            <PageHeader
                icon={PuzzleOutlineIcon}
                title="Header"
                actions={
                    <Link to={`${location.pathname}/close`} className="btn btn-secondary mr-1">
                        <SearchIcon className="icon-inline" /> Button with icon
                    </Link>
                }
            />
        )}
    </WebStory>
))

add('Complex header', () => (
    <WebStory>
        {() => (
            <PageHeader
                annotation={<Link to="/page">Share feedback</Link>}
                title={
                    <>
                        <Link to="/level-1">
                            <PuzzleOutlineIcon className="icon-inline" />
                            Level 1
                        </Link>{' '}
                        / <Link to="/level-2">Level 2</Link> / Page name
                    </>
                }
                byline="Created 3 months ago"
                actions={
                    <Link to="/page" className="btn btn-secondary mr-1">
                        <SearchIcon className="icon-inline" /> Button with icon
                    </Link>
                }
            />
        )}
    </WebStory>
))
