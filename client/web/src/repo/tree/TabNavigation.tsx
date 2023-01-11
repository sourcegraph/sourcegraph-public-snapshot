import React from 'react'

import { mdiSourceCommit, mdiSourceBranch, mdiTag, mdiHistory, mdiAccount, mdiBrain, mdiCog } from '@mdi/js'

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
            <Icon aria-hidden={true} svgPath={mdiSourceCommit} /> Commits
        </Button>
        <Button onClick={() => setCurrentTab('branches')} variant="secondary" outline={true}>
            <Icon aria-hidden={true} svgPath={mdiSourceBranch} /> Branches
        </Button>
        <Button onClick={() => setCurrentTab('tags')} variant="secondary" outline={true}>
            <Icon aria-hidden={true} svgPath={mdiTag} /> Tags
        </Button>
        <Button onClick={() => setCurrentTab('compare')} variant="secondary" outline={true}>
            <Icon aria-hidden={true} svgPath={mdiHistory} /> Compare
        </Button>
        <Button onClick={() => setCurrentTab('contributors')} variant="secondary" outline={true}>
            <Icon aria-hidden={true} svgPath={mdiAccount} /> Contributors
        </Button>
        {codeIntelligenceEnabled && (
            <Button
                to={`/${encodeURIPathComponent(repo.name)}/-/code-graph`}
                variant="secondary"
                outline={true}
                as={Link}
            >
                <Icon aria-hidden={true} svgPath={mdiBrain} /> Code graph data
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
                <Icon aria-hidden={true} svgPath={mdiCog} /> Settings
            </Button>
        )}
    </ButtonGroup>
)
