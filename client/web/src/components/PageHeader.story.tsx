import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../SourcegraphWebApp.scss'
import { PageHeader } from './PageHeader'
import { Link } from '../../../shared/src/components/Link'
import SearchIcon from 'mdi-react/SearchIcon'
import { CampaignsIconFlushLeft } from '../enterprise/campaigns/icons'
import { WebStory } from './WebStory'

const { add } = storiesOf('web/PageHeader', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container mt-3">{story()}</div>
        </div>
    </>
))

// TODO: CLEAN UP EXAMPLES

add('Basic header', () => (
    <WebStory>
        {() => (
            <PageHeader
                icon={CampaignsIconFlushLeft}
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
                annotation={<Link to="/thing">Share feedback</Link>}
                title={
                    <>
                        <Link to="/anywhere">
                            <CampaignsIconFlushLeft className="icon-inline" />
                            Level 1
                        </Link>{' '}
                        / <Link to="/somewhere">Level 2</Link> / Page name
                    </>
                }
                subtitle="Created 3 months ago"
                actions={
                    <Link to={`${location.pathname}/close`} className="btn btn-secondary mr-1">
                        <SearchIcon className="icon-inline" /> Button with icon
                    </Link>
                }
            />
        )}
    </WebStory>
))
