import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createThread } from '../../../discussions/backend'
import { computeDiff } from '../../threads/detail/changes/computeDiff'
import { ThreadSettings } from '../../threads/settings'

const FAKE_PROJECT_ID = 'UHJvamVjdDox' // TODO!(sqs)

/**
 * Create a preview changeset by applying the {@link codeAction}.
 */
export async function createPreviewChangeset(
    { extensionsController }: ExtensionsControllerProps,
    codeAction: sourcegraph.CodeAction
): Promise<Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'url'>> {
    const settings: ThreadSettings = { previewChangesetDiff: await computeDiff(extensionsController, [codeAction]) }
    return createThread({
        type: GQL.ThreadType.CHANGESET,
        title: codeAction.title,
        contents: '',
        project: FAKE_PROJECT_ID,
        settings: JSON.stringify(settings, null, 2),
        status: GQL.ThreadStatus.PREVIEW,
    }).toPromise()
}
