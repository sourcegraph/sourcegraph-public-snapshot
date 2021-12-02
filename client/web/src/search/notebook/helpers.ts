import { serializeBlockInput } from './serialize'

import { BlockInput } from '.'

export function serializeBlocks(blocks: BlockInput[]): string {
    return blocks
        .map(
            block =>
                `${encodeURIComponent(block.type)}:${encodeURIComponent(
                    serializeBlockInput(block, window.location.origin)
                )}`
        )
        .join(',')
}
