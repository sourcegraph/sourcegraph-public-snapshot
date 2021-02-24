import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../main.scss'
import { SiteAdminSidebar } from './SiteAdminSidebar'
import EarthIcon from 'mdi-react/EarthIcon'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import { MemoryRouter } from 'react-router'
import { enterpriseSiteAdminSidebarGroups } from '../enterprise/site-admin/sidebaritems'

const groups = [
    {
        header: {
            label: 'With icon and collapsible',
            icon: EarthIcon,
        },
        items: [
            {
                label: 'Overview',
                to: '/site-admin',
                exact: true,
            },
            {
                label: 'Usage stats',
                to: '/site-admin/usage-statistics',
            },
            {
                label: 'Feedback survey',
                to: '/site-admin/surveys',
            },
        ],
    },
    {
        header: {
            label: 'Without icon and collapsible',
        },
        items: [
            {
                label: 'Overview',
                to: '/site-admin',
                exact: true,
            },
            {
                label: 'Usage stats',
                to: '/site-admin/usage-statistics',
            },
            {
                label: 'Feedback survey',
                to: '/site-admin/surveys',
            },
        ],
    },
    {
        header: {
            label: 'With icon and non-collapsible',
            icon: PuzzleOutlineIcon,
        },
        items: [
            {
                label: 'Extensions',
                to: '/some/extensions',
            },
        ],
    },
    {
        header: {
            label: 'Without icon and non-collapsible',
        },
        items: [
            {
                label: 'Business',
                to: '/some/business',
            },
        ],
    },
]

const { add } = storiesOf('web/site-admin/AdminSidebar', module).addDecorator(story => (
    <>
        <div>{story()}</div>
        <style>{webStyles}</style>
    </>
))

add(
    'Collapsible',
    () => (
        <MemoryRouter>
            <SiteAdminSidebar className={'site-admin-sidebar'} groups={[groups[1]]} />
        </MemoryRouter>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/cT23UcGqbdjdV4H2yxweIu/%2311775-Map-the-current-information-architecture-%5BApproved%5D?node-id=0%3A1',
        },
    }
)

add(
    'Collapsible with icon',
    () => (
        <MemoryRouter>
            <SiteAdminSidebar className={'site-admin-sidebar'} groups={[groups[0]]} />
        </MemoryRouter>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/cT23UcGqbdjdV4H2yxweIu/%2311775-Map-the-current-information-architecture-%5BApproved%5D?node-id=0%3A1',
        },
    }
)

add(
    'Non-collapsible with icon',
    () => (
        <MemoryRouter>
            <SiteAdminSidebar className={'site-admin-sidebar'} groups={[groups[2]]} />
        </MemoryRouter>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/cT23UcGqbdjdV4H2yxweIu/%2311775-Map-the-current-information-architecture-%5BApproved%5D?node-id=0%3A1',
        },
    }
)

add(
    'Non-collapsible',
    () => (
        <MemoryRouter>
            <SiteAdminSidebar className={'site-admin-sidebar'} groups={[groups[3]]} />
        </MemoryRouter>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/cT23UcGqbdjdV4H2yxweIu/%2311775-Map-the-current-information-architecture-%5BApproved%5D?node-id=0%3A1',
        },
    }
)

add(
    'Dropdown and single link',
    () => (
        <MemoryRouter>
            <SiteAdminSidebar className={'site-admin-sidebar'} groups={groups} />
        </MemoryRouter>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/cT23UcGqbdjdV4H2yxweIu/%2311775-Map-the-current-information-architecture-%5BApproved%5D?node-id=0%3A1',
        },
    }
)

add(
    'Enterprise Items',
    () => (
        <MemoryRouter>
            <SiteAdminSidebar className={'site-admin-sidebar'} groups={enterpriseSiteAdminSidebarGroups} />
        </MemoryRouter>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/cT23UcGqbdjdV4H2yxweIu/%2311775-Map-the-current-information-architecture-%5BApproved%5D?node-id=0%3A1',
        },
    }
)
