import BranchIcon from '@sourcegraph/icons/lib/Branch'
import GitHubIcon from '@sourcegraph/icons/lib/GitHub'
import TagIcon from '@sourcegraph/icons/lib/Tag'
import React from 'react'

interface Props {
    gitRef: GQL.IGitRef
}

export const GitRefTag: React.SFC<Props> = ({ gitRef }: Props) => {
    // TODO(sqs): make not github specific
    const githubRepoURL = gitRef.repository.uri.startsWith('github.com/')
        ? `https://${gitRef.repository.uri}`
        : undefined

    const abbrevName = gitRef.name.slice(gitRef.prefix.length)
    let kind = ''
    let url = githubRepoURL || ''
    let Icon: React.ComponentType<{ className: string }> | undefined
    switch (gitRef.prefix) {
        case 'refs/heads/':
            kind = 'branch'
            url = url && `${url}/compare/${abbrevName}`
            Icon = BranchIcon
            break

        case 'refs/tags/':
            kind = 'tag'
            url = url && `${url}/releases/tag/${abbrevName}`
            Icon = TagIcon
            break

        case 'refs/pull/':
            if (gitRef.name.endsWith('/head')) {
                kind = 'pull request'
                url = url && `${url}/pull/${abbrevName.split('/')[0]}`
                Icon = GitHubIcon
            } else if (gitRef.name.endsWith('/merge')) {
                kind = 'pull request merge'
                url = url && `${url}/pull/${abbrevName.split('/')[0]}`
                Icon = GitHubIcon
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
        <a href={url} {...props}>
            {children}
        </a>
    ) : (
        <span {...props}>{children}</span>
    )
}
