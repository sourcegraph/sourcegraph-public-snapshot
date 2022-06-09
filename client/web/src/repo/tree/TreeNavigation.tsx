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

interface TreeNavigationProps {
    repo: TreePageRepositoryFields
    revision: string
    tree: TreeFields
    codeIntelligenceEnabled: boolean
    batchChangesEnabled: boolean
}

export const TreeNavigation: React.FunctionComponent<React.PropsWithChildren<TreeNavigationProps>> = ({
    repo,
    revision,
    tree,
    codeIntelligenceEnabled,
    batchChangesEnabled,
}) => (
    <ButtonGroup>
        <Button to={`${tree.url}/-/commits`} variant="secondary" outline={true} as={Link}>
            <Icon role="img" as={SourceCommitIcon} aria-hidden={true} /> Commits
        </Button>
        <Button to={`/${encodeURIPathComponent(repo.name)}/-/branches`} variant="secondary" outline={true} as={Link}>
            <Icon role="img" as={SourceBranchIcon} aria-hidden={true} /> Branches
        </Button>
        <Button to={`/${encodeURIPathComponent(repo.name)}/-/tags`} variant="secondary" outline={true} as={Link}>
            <Icon role="img" as={TagIcon} aria-hidden={true} /> Tags
        </Button>
        <Button
            to={
                revision
                    ? `/${encodeURIPathComponent(repo.name)}/-/compare/...${encodeURIComponent(revision)}`
                    : `/${encodeURIPathComponent(repo.name)}/-/compare`
            }
            variant="secondary"
            outline={true}
            as={Link}
        >
            <Icon role="img" as={HistoryIcon} aria-hidden={true} /> Compare
        </Button>
        <Button
            to={`/${encodeURIPathComponent(repo.name)}/-/stats/contributors`}
            variant="secondary"
            outline={true}
            as={Link}
        >
            <Icon role="img" as={AccountIcon} aria-hidden={true} /> Contributors
        </Button>
        {codeIntelligenceEnabled && (
            <Button
                to={`/${encodeURIPathComponent(repo.name)}/-/code-intelligence`}
                variant="secondary"
                outline={true}
                as={Link}
            >
                <Icon role="img" as={BrainIcon} aria-hidden={true} /> Code Intelligence
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
                <Icon role="img" as={SettingsIcon} aria-hidden={true} /> Settings
            </Button>
        )}
    </ButtonGroup>
)
