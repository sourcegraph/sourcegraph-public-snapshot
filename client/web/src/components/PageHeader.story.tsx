import { storiesOf } from '@storybook/react'
import React from 'react'
import SearchIcon from 'mdi-react/SearchIcon'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'

import webStyles from '../SourcegraphWebApp.scss'
import { PageHeader } from './PageHeader'
import { PageBreadcrumbs } from './PageBreadcrumbs'
import { Link } from '../../../shared/src/components/Link'
import { StatusAnnotation } from './StatusAnnotation'

const { add } = storiesOf('web/PageHeader', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container web-content mt-3">{story()}</div>
        </div>
    </>
))

add('Basic header', () => (
    <PageHeader
        icon={PuzzleOutlineIcon}
        title="Header"
        actions={
            <Link to={`${location.pathname}/close`} className="btn btn-secondary mr-1">
                <SearchIcon className="icon-inline" /> Button with icon
            </Link>
        }
    />
))

add('Complex header', () => (
    <PageHeader
        annotation={<StatusAnnotation status="beta" />}
        title={
            <PageBreadcrumbs
                icon={PuzzleOutlineIcon}
                path={[{ to: '/level-1', text: 'Level 1' }, { text: 'Level 2' }]}
            />
        }
        byline={
            <>
                Created by <Link to="/page">user</Link> 3 months ago
            </>
        }
        actions={
            <Link to="/page" className="btn btn-secondary mr-1">
                <SearchIcon className="icon-inline" /> Button with icon
            </Link>
        }
    />
))
