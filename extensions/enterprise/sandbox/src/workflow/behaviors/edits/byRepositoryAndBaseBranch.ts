import * as sourcegraph from 'sourcegraph'
import { EditsBehaviorCommandContext, Changeset } from '.'
import { parseRepoURI } from '../../../../../../../shared/src/util/url'
import { queryGraphQL } from '../../../util'
import { createAggregateError } from '../../../../../../../shared/src/util/errors'

const repositoryInfoByName = async (repositoryName: string): Promise<{ id: string; defaultBranch: string }> => {
    const { data, errors } = await queryGraphQL({
        query: `
	query RepositoryIDByName($repositoryName: String!) {
		repository(name: $repositoryName) {
            id
            defaultBranch {
                name
            }
		}
	}`,
        vars: { repositoryName },
    })
    if (errors && errors.length > 0) {
        throw createAggregateError(errors)
    }
    if (!data || !data.repository) {
        throw new Error(`no repository with name ${JSON.stringify(repositoryName)}`)
    }
    if (!data.repository.defaultBranch) {
        throw new Error(`no default branch for repository ${JSON.stringify(repositoryName)}`)
    }
    return { id: data.repository.id, defaultBranch: data.repository.defaultBranch.name }
}

const changesetsByRepositoryAndBaseBranch = async (
    edit: sourcegraph.WorkspaceEdit,
    context: EditsBehaviorCommandContext
): Promise<Changeset[]> => {
    const changesets = new Map<string, Changeset>()
    for (const [uri, edits] of edit.textEdits()) {
        const p = parseRepoURI(uri.toString())

        let e = changesets.get(p.repoName)
        if (!e) {
            const { id: repoID, defaultBranch } = await repositoryInfoByName(p.repoName)
            e = {
                title: context.title,
                body: context.body,
                baseRepository: repoID,
                baseBranch: defaultBranch,
                headRepository: repoID,
                headBranch: context.headBranch,
                edit: new sourcegraph.WorkspaceEdit(),
            }
            changesets.set(p.repoName, e)
        }
        e.edit.set(uri, edits)
    }
    return Array.from(changesets.values()).map(({ edit, ...changeset }) => ({
        ...changeset,
        edit,
    }))
}

export const CHANGESET_BEHAVIOR_BY_REPOSITORY_AND_BASE_BRANCH_COMMAND = [
    'changesets.byRepositoryAndBaseBranch',
    changesetsByRepositoryAndBaseBranch,
] as const
