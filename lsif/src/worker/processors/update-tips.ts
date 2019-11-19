import { TracingContext } from '../../shared/tracing'
import { XrepoDatabase } from '../../shared/xrepo/xrepo'

/**
 * Create a job that updates the tip of the default branch for every repository that has LSIF data.
 *
 * @param xrepoDatabase The cross-repo database.
 * @param fetchConfiguration A function that returns the current configuration.
 */
export const createUpdateTipsJobProcessor = (
    xrepoDatabase: XrepoDatabase,
    fetchConfiguration: () => { gitServers: string[] }
) => (_: unknown, ctx: TracingContext): Promise<void> =>
    xrepoDatabase.discoverAndUpdateTips({
        gitserverUrls: fetchConfiguration().gitServers,
        ctx,
    })
