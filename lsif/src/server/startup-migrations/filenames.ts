import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as path from 'path'
import * as settings from '../settings'
import { dbFilename, idFromFilename } from '../../shared/paths'
import { TracingContext, logAndTraceCall } from '../../shared/tracing'

/**
 * If it hasn't been done already, migrate from the old pre-3.13 filename format
 * `$ID-$REPO@$COMMIT.lsif.db` to the new format `$ID.lsif.db`.
 *
 * @param ctx The tracing context.
 */
export function migrateFilenames(ctx: TracingContext): Promise<void> {
    return logAndTraceCall(ctx, 'Migrating database filenames', async () => {
        const doneFile = path.join(settings.STORAGE_ROOT, 'id-only-based-filenames')
        if (await fs.exists(doneFile)) {
            // Already migrated.
            return
        }

        for (const basename of await fs.readdir(path.join(settings.STORAGE_ROOT, constants.DBS_DIR))) {
            const id = idFromFilename(basename)
            if (!id) {
                continue
            }

            await fs.rename(
                path.join(settings.STORAGE_ROOT, constants.DBS_DIR, basename),
                dbFilename(settings.STORAGE_ROOT, id)
            )
        }

        // Create an empty done file to record that all files have been renamed.
        await fs.close(await fs.open(doneFile, 'w'))
    })
}
