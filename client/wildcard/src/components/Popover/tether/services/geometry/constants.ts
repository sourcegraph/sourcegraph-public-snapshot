import { createPoint } from '../../models/geometry/point'
import { Position, Side } from '../../models/tether-models'

/**
 * Static position preferences settings for each possible pre-defined position.
 */
export const POSITION_VARIANTS = {
    [Position.topLeft]: {
        positionSides: Side.top,
        rotationAngle: 0,
        opposite: Position.bottomLeft,
        elementAttachments: createPoint(0, 1),
        targetAttachments: createPoint(0, 0),
        targetOffset: createPoint(0, 1),
    },
    [Position.topCenter]: {
        positionSides: Side.top,
        rotationAngle: 0,
        opposite: Position.bottomCenter,
        elementAttachments: createPoint(0.5, 1),
        targetAttachments: createPoint(0.5, 0),
        targetOffset: createPoint(0, 1),
    },
    [Position.topRight]: {
        positionSides: Side.top,
        rotationAngle: 0,
        opposite: Position.bottomRight,
        elementAttachments: createPoint(1, 1),
        targetAttachments: createPoint(1, 0),
        targetOffset: createPoint(0, 1),
    },
    [Position.rightTop]: {
        positionSides: Side.right,
        rotationAngle: 90,
        opposite: Position.leftTop,
        elementAttachments: createPoint(0, 0),
        targetAttachments: createPoint(1, 0),
        targetOffset: createPoint(1, 0),
    },
    [Position.rightMiddle]: {
        positionSides: Side.right,
        rotationAngle: 90,
        opposite: Position.leftMiddle,
        elementAttachments: createPoint(0, 0.5),
        targetAttachments: createPoint(1, 0.5),
        targetOffset: createPoint(1, 0),
    },
    [Position.rightBottom]: {
        positionSides: Side.right,
        rotationAngle: 90,
        opposite: Position.leftBottom,
        elementAttachments: createPoint(0, 1),
        targetAttachments: createPoint(1, 1),
        targetOffset: createPoint(1, 0),
    },
    [Position.bottomLeft]: {
        positionSides: Side.bottom,
        rotationAngle: 180,
        opposite: Position.topLeft,
        elementAttachments: createPoint(0, 0),
        targetAttachments: createPoint(0, 1),
        targetOffset: createPoint(0, 1),
    },
    [Position.bottomCenter]: {
        positionSides: Side.bottom,
        rotationAngle: 180,
        opposite: Position.topCenter,
        elementAttachments: createPoint(0.5, 0),
        targetAttachments: createPoint(0.5, 1),
        targetOffset: createPoint(0, 1),
    },
    [Position.bottomRight]: {
        positionSides: Side.bottom,
        rotationAngle: 180,
        opposite: Position.topRight,
        elementAttachments: createPoint(1, 0),
        targetAttachments: createPoint(1, 1),
        targetOffset: createPoint(0, 1),
    },
    [Position.leftTop]: {
        positionSides: Side.left,
        rotationAngle: 270,
        opposite: Position.rightTop,
        elementAttachments: createPoint(1, 0),
        targetAttachments: createPoint(0, 0),
        targetOffset: createPoint(1, 0),
    },
    [Position.leftMiddle]: {
        positionSides: Side.left,
        rotationAngle: 270,
        opposite: Position.rightMiddle,
        elementAttachments: createPoint(1, 0.5),
        targetAttachments: createPoint(0, 0.5),
        targetOffset: createPoint(1, 0),
    },
    [Position.leftBottom]: {
        positionSides: Side.left,
        rotationAngle: 270,
        opposite: Position.rightBottom,
        elementAttachments: createPoint(1, 1),
        targetAttachments: createPoint(0, 1),
        targetOffset: createPoint(1, 0),
    },
}
