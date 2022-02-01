import { createPoint } from '../../models/geometry/point'
import { Position } from '../../models/tether-models'

/**
 * Static position preferences settings for each possible pre-defined position.
 */
export const POSITION_VARIANTS = {
    [Position.topStart]: {
        positionSides: Position.top,
        rotationAngle: 0,
        opposite: Position.bottomStart,
        elementAttachments: createPoint(0, 1),
        targetAttachments: createPoint(0, 0),
        targetOffset: createPoint(0, 1),
    },
    [Position.top]: {
        positionSides: Position.top,
        rotationAngle: 0,
        opposite: Position.bottom,
        elementAttachments: createPoint(0.5, 1),
        targetAttachments: createPoint(0.5, 0),
        targetOffset: createPoint(0, 1),
    },
    [Position.topEnd]: {
        positionSides: Position.top,
        rotationAngle: 0,
        opposite: Position.bottomEnd,
        elementAttachments: createPoint(1, 1),
        targetAttachments: createPoint(1, 0),
        targetOffset: createPoint(0, 1),
    },
    [Position.rightStart]: {
        positionSides: Position.right,
        rotationAngle: 90,
        opposite: Position.leftStart,
        elementAttachments: createPoint(0, 0),
        targetAttachments: createPoint(1, 0),
        targetOffset: createPoint(1, 0),
    },
    [Position.right]: {
        positionSides: Position.right,
        rotationAngle: 90,
        opposite: Position.left,
        elementAttachments: createPoint(0, 0.5),
        targetAttachments: createPoint(1, 0.5),
        targetOffset: createPoint(1, 0),
    },
    [Position.rightEnd]: {
        positionSides: Position.right,
        rotationAngle: 90,
        opposite: Position.leftEnd,
        elementAttachments: createPoint(0, 1),
        targetAttachments: createPoint(1, 1),
        targetOffset: createPoint(1, 0),
    },
    [Position.bottomStart]: {
        positionSides: Position.bottom,
        rotationAngle: 180,
        opposite: Position.topStart,
        elementAttachments: createPoint(0, 0),
        targetAttachments: createPoint(0, 1),
        targetOffset: createPoint(0, 1),
    },
    [Position.bottom]: {
        positionSides: Position.bottom,
        rotationAngle: 180,
        opposite: Position.top,
        elementAttachments: createPoint(0.5, 0),
        targetAttachments: createPoint(0.5, 1),
        targetOffset: createPoint(0, 1),
    },
    [Position.bottomEnd]: {
        positionSides: Position.bottom,
        rotationAngle: 180,
        opposite: Position.topEnd,
        elementAttachments: createPoint(1, 0),
        targetAttachments: createPoint(1, 1),
        targetOffset: createPoint(0, 1),
    },
    [Position.leftStart]: {
        positionSides: Position.left,
        rotationAngle: 270,
        opposite: Position.rightStart,
        elementAttachments: createPoint(1, 0),
        targetAttachments: createPoint(0, 0),
        targetOffset: createPoint(1, 0),
    },
    [Position.left]: {
        positionSides: Position.left,
        rotationAngle: 270,
        opposite: Position.right,
        elementAttachments: createPoint(1, 0.5),
        targetAttachments: createPoint(0, 0.5),
        targetOffset: createPoint(1, 0),
    },
    [Position.leftEnd]: {
        positionSides: Position.left,
        rotationAngle: 270,
        opposite: Position.rightEnd,
        elementAttachments: createPoint(1, 1),
        targetAttachments: createPoint(0, 1),
        targetOffset: createPoint(1, 0),
    },
} as const
