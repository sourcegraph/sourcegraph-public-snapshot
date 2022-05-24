import React from 'react'

import AccountIcon from 'mdi-react/AccountIcon'
import BrainIcon from 'mdi-react/BrainIcon'
import HistoryIcon from 'mdi-react/HistoryIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import TagIcon from 'mdi-react/TagIcon'

import { encodeURIPathComponent } from '@sourcegraph/common'
import { TreeFields } from '@sourcegraph/shared/src/graphql-operations'
import { Button, ButtonGroup, Icon, Link } from '@sourcegraph/wildcard'

import { RepoBatchChangesButton } from '../../batches/RepoBatchChangesButton'
import { TreePageRepositoryFields } from '../../graphql-operations'

interface TabNavigationProps {
    setCurrentTab(tabName: string): (tabName: string) => {}
    repo: TreePageRepositoryFields
    revision: string
    tree: TreeFields
    codeIntelligenceEnabled: boolean
    batchChangesEnabled: boolean
}

export const TabNavigation: React.FunctionComponent<React.PropsWithChildren<TabNavigationProps>> = ({
    setCurrentTab,
    repo,
    codeIntelligenceEnabled,
    batchChangesEnabled,
}) => (
    <ButtonGroup>
        <Button onClick={() => setCurrentTab('commits')} variant="secondary" outline={true}>
            <Icon as={SourceCommitIcon} /> Commits
        </Button>
        <Button onClick={() => setCurrentTab('branches')} variant="secondary" outline={true}>
            <Icon as={SourceBranchIcon} /> Branches
        </Button>
        <Button onClick={() => setCurrentTab('tags')} variant="secondary" outline={true}>
            <Icon as={TagIcon} /> Tags
        </Button>
        <Button onClick={() => setCurrentTab('compare')} variant="secondary" outline={true}>
            <Icon as={HistoryIcon} /> Compare
        </Button>
        <Button onClick={() => setCurrentTab('contributors')} variant="secondary" outline={true}>
            <Icon as={AccountIcon} /> Contributors
        </Button>
        {codeIntelligenceEnabled && (
            <Button
                to={`/${encodeURIPathComponent(repo.name)}/-/code-intelligence`}
                variant="secondary"
                outline={true}
                as={Link}
            >
                <Icon as={BrainIcon} /> Code Intelligence
            </Button>
        )}
        {batchChangesEnabled && <RepoBatchChangesButton repoName={repo.name} />}
        {repo.viewerCanAdminister && (
            <Button
                to={`/${encodeURIPathComponent(repo.name)}/-/settings`}
                variant="secondary"
                outline={true}
                as={Link}
            >
                <Icon as={SettingsIcon} /> Settings
            </Button>
        )}
    </ButtonGroup>
)
