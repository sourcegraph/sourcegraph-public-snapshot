import * as prettier from 'prettier'
import { SourceFile } from 'ts-morph'

/**
 * Formats the given source file with Prettier.
 */
export async function formatSourceFile(sourceFile: SourceFile): Promise<SourceFile> {
    const config = await prettier.resolveConfig(sourceFile.getFilePath())
    if (!config) {
        throw new Error(`Prettier config not found for file ${sourceFile.getFilePath() as string}`)
    }
    return sourceFile.replaceWithText(
        prettier.format(sourceFile.getFullText(), {
            ...config,
            filepath: sourceFile.getFilePath(),
        })
    ) as SourceFile
}
