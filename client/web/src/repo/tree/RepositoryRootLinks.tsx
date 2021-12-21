import AccountIcon from 'mdi-react/AccountIcon'
import BookOpenBlankVariantIcon from 'mdi-react/BookOpenBlankVariantIcon'
import BrainIcon from 'mdi-react/BrainIcon'
import HistoryIcon from 'mdi-react/HistoryIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import TagIcon from 'mdi-react/TagIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { encodeURIPathComponent, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { BatchChangesProps } from '../../batches'
import { RepoBatchChangesButton } from '../../batches/RepoBatchChangesButton'
import { CodeIntelligenceProps } from '../../codeintel'
import { TreeFields, TreePageRepositoryFields } from '../../graphql-operations'
import { useExperimentalFeatures } from '../../stores'

interface Props extends RevisionSpec, CodeIntelligenceProps, Pick<BatchChangesProps, 'batchChangesEnabled'> {
    repo: TreePageRepositoryFields
    tree: TreeFields
}

export const RepositoryRootLinks: React.FunctionComponent<Props> = ({
    repo,
    tree,
    revision,
    codeIntelligenceEnabled,
    batchChangesEnabled,
}) => {
    // eslint-disable-next-line unicorn/prevent-abbreviations
    const enableAPIDocs = useExperimentalFeatures(features => features.apiDocs)

    return (
        <div className="btn-group">
            {enableAPIDocs && (
                <Link className="btn btn-outline-secondary" to={`${tree.url}/-/docs`}>
                    <BookOpenBlankVariantIcon className="icon-inline" /> API docs
                </Link>
            )}
            <Link className="btn btn-outline-secondary" to={`${tree.url}/-/commits`}>
                <SourceCommitIcon className="icon-inline" /> Commits
            </Link>
            <Link className="btn btn-outline-secondary" to={`/${encodeURIPathComponent(repo.name)}/-/branches`}>
                <SourceBranchIcon className="icon-inline" /> Branches
            </Link>
            <Link className="btn btn-outline-secondary" to={`/${encodeURIPathComponent(repo.name)}/-/tags`}>
                <TagIcon className="icon-inline" /> Tags
            </Link>
            <Link
                className="btn btn-outline-secondary"
                to={
                    revision
                        ? `/${encodeURIPathComponent(repo.name)}/-/compare/...${encodeURIComponent(revision)}`
                        : `/${encodeURIPathComponent(repo.name)}/-/compare`
                }
            >
                <HistoryIcon className="icon-inline" /> Compare
            </Link>
            <Link
                className="btn btn-outline-secondary"
                to={`/${encodeURIPathComponent(repo.name)}/-/stats/contributors`}
            >
                <AccountIcon className="icon-inline" /> Contributors
            </Link>
            {codeIntelligenceEnabled && (
                <Link
                    className="btn btn-outline-secondary"
                    to={`/${encodeURIPathComponent(repo.name)}/-/code-intelligence`}
                >
                    <BrainIcon className="icon-inline" /> Code Intelligence
                </Link>
            )}
            {batchChangesEnabled && (
                <RepoBatchChangesButton className="btn btn-outline-secondary" repoName={repo.name} />
            )}
            {repo.viewerCanAdminister && (
                <Link className="btn btn-outline-secondary" to={`/${encodeURIPathComponent(repo.name)}/-/settings`}>
                    <SettingsIcon className="icon-inline" /> Settings
                </Link>
            )}
        </div>
    )
}
