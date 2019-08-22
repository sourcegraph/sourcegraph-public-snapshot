import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import TagIcon from 'mdi-react/TagIcon'
import React from 'react'
import * as GQL from '../../../shared/src/graphql/schema'

interface Props {
    gitRef: GQL.IGitRef

    /**
     * Called when the mousedown event is triggered on the element.
     */
    onMouseDown?: () => void
}

export const GitRefTag: React.FunctionComponent<Props> = ({ gitRef, onMouseDown }: Props) => {
    // TODO(sqs): make not github specific
    const githubRepoURL = gitRef.repository.name.startsWith('github.com/')
        ? `https://${gitRef.repository.name}`
        : undefined

    const abbrevName = gitRef.name.slice(gitRef.prefix.length)
    let kind = ''
    let url = githubRepoURL || ''
    let Icon: React.ComponentType<{ className?: string }> | undefined
    switch (gitRef.prefix) {
        case 'refs/heads/':
            kind = 'branch'
            url = url && `${url}/compare/${encodeURIComponent(abbrevName)}`
            Icon = SourceBranchIcon
            break

        case 'refs/tags/':
            kind = 'tag'
            url = url && `${url}/releases/tag/${encodeURIComponent(abbrevName)}`
            Icon = TagIcon
            break

        case 'refs/pull/':
            if (gitRef.name.endsWith('/head')) {
                kind = 'pull request'
                url = url && `${url}/pull/${abbrevName.split('/')[0]}`
                Icon = GithubCircleIcon
            } else if (gitRef.name.endsWith('/merge')) {
                kind = 'pull request merge'
                url = url && `${url}/pull/${abbrevName.split('/')[0]}`
                Icon = GithubCircleIcon
            }
            break
    }

    // Render <a> if there's a canonical URL for the Git ref. Otherwise render <span>.
    const children: (React.ReactChild | undefined)[] = [
        Icon && <Icon key={1} className="icon-inline" />,
        <span key={2} className="git-ref-tag__display-name">
            {gitRef.displayName}
        </span>,
    ]
    const props = {
        title: `${gitRef.name} ${kind ? `(${kind})` : ''}`,
        className: 'git-ref-tag',
    }

    return url ? (
        <a href={url} {...props} onMouseDown={onMouseDown}>
            {children}
        </a>
    ) : (
        <span {...props}>{children}</span>
    )
}
