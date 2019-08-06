import BellIcon from 'mdi-react/BellIcon'
import UserGroupIcon from 'mdi-react/UserGroupIcon'
import UserIcon from 'mdi-react/UserIcon'
import React from 'react'
import { Toggle } from '../../../../../shared/src/components/Toggle'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { InfoSidebarSection } from '../../../components/infoSidebar/InfoSidebar'
import { LabelIcon } from '../../../projects/icons'
import { CampaignsIcon } from '../../campaigns/icons'
import { ObjectCampaignsList } from '../../campaigns/object/ObjectCampaignsList'
import { CopyThreadLinkButton } from '../../threadsOLD/detail/CopyThreadLinkButton'
import { ThreadCampaignsDropdownButton } from '../ThreadCampaignsDropdownButton'
import { ThreadStateBadge } from '../threadState/ThreadStateBadge'

interface Props extends ExtensionsControllerNotificationProps {
    thread: GQL.ThreadOrIssueOrChangeset
    onThreadUpdate: () => void
}

export const threadlikeSidebarSections = ({ thread, onThreadUpdate, ...props }: Props): InfoSidebarSection[] => [
    {
        expanded: <ThreadStateBadge thread={thread} />,
        collapsed: <ThreadStateBadge thread={thread} />,
    },
    {
        expanded: {
            title: (
                <ThreadCampaignsDropdownButton
                    {...props}
                    thread={thread}
                    onChange={onThreadUpdate}
                    buttonClassName="btn-link p-0"
                />
            ),
            children: <ObjectCampaignsList object={thread} icon={false} itemClassName="small" />,
        },
        collapsed: {
            icon: CampaignsIcon,
            tooltip: 'Campaign',
        },
    },
    {
        expanded: {
            title: 'Assignee',
            children: <strong>@sqs</strong>,
        },
        collapsed: {
            icon: UserIcon,
            tooltip: 'Assignee: @sqs',
        },
    },
    {
        expanded: {
            title: 'Labels',
            children: thread.title
                .toLowerCase()
                .split(' ')
                .filter(w => w.length >= 3)
                .map((label, i) => (
                    <span key={i} className={`badge badge-secondary mr-1`}>
                        {label}
                    </span>
                )),
        },
        collapsed: {
            icon: LabelIcon,
            tooltip: 'Labels',
        },
    },
    {
        expanded: {
            title: '3 participants',
            children: <div className="text-muted">@sqs @jtal3sf @xyzhao</div>,
        },
        collapsed: {
            icon: UserGroupIcon,
            tooltip: '3 participants',
        },
    },
    {
        expanded: {
            title: (
                <div className="d-flex align-items-center justify-content-between">
                    Notifications <Toggle value={true} />
                </div>
            ),
        },
        collapsed: {
            icon: BellIcon,
            tooltip: 'Notifications: on',
        },
    },
    {
        expanded: {
            title: (
                <div className="d-flex align-items-center justify-content-between">
                    Link{' '}
                    <CopyThreadLinkButton
                        link={'aasdf TODO!(sqs)'}
                        className="btn btn-link btn-link-sm text-decoration-none px-0"
                    >
                        #{thread.number}
                    </CopyThreadLinkButton>
                </div>
            ),
        },
        collapsed: (
            <CopyThreadLinkButton
                link={'aasdf TODO!(sqs)'}
                className="btn btn-link btn-link-sm text-decoration-none px-0"
            >
                #{thread.number}
            </CopyThreadLinkButton>
        ),
    },
]
