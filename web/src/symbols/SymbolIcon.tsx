import CodeArrayIcon from '@sourcegraph/icons/lib/CodeArray'
import CodeBracesIcon from '@sourcegraph/icons/lib/CodeBraces'
import CodeNotEqualIcon from '@sourcegraph/icons/lib/CodeNotEqual'
import CodeStringIcon from '@sourcegraph/icons/lib/CodeString'
import CubeIcon from '@sourcegraph/icons/lib/Cube'
import CubeOutlineIcon from '@sourcegraph/icons/lib/CubeOutline'
import DrawingBoxIcon from '@sourcegraph/icons/lib/DrawingBox'
import FileDocumentIcon from '@sourcegraph/icons/lib/FileDocument'
import FunctionIcon from '@sourcegraph/icons/lib/Function'
import KeyIcon from '@sourcegraph/icons/lib/Key'
import LinkIcon from '@sourcegraph/icons/lib/Link'
import MatrixIcon from '@sourcegraph/icons/lib/Matrix'
import NullIcon from '@sourcegraph/icons/lib/Null'
import NumericIcon from '@sourcegraph/icons/lib/Numeric'
import PackageIcon from '@sourcegraph/icons/lib/Package'
import PiBoxIcon from '@sourcegraph/icons/lib/PiBox'
import PillarIcon from '@sourcegraph/icons/lib/Pillar'
import PoundIcon from '@sourcegraph/icons/lib/Pound'
import ShapeIcon from '@sourcegraph/icons/lib/Shape'
import SitemapIcon from '@sourcegraph/icons/lib/Sitemap'
import TextboxIcon from '@sourcegraph/icons/lib/Textbox'
import TimetableIcon from '@sourcegraph/icons/lib/Timetable'
import WebIcon from '@sourcegraph/icons/lib/Web'
import WrenchIcon from '@sourcegraph/icons/lib/Wrench'
import { kebabCase } from 'lodash'
import * as React from 'react'

/**
 * Returns the icon component for a given symbol kind
 */
const getSymbolIconComponent = (kind: GQL.SymbolKind): React.ComponentType<React.HTMLAttributes<HTMLElement>> => {
    switch (kind) {
        case 'FILE':
            return FileDocumentIcon
        case 'MODULE':
            return CodeBracesIcon
        case 'NAMESPACE':
            return WebIcon
        case 'PACKAGE':
            return PackageIcon
        case 'CLASS':
            return SitemapIcon
        case 'METHOD':
            return CubeOutlineIcon
        case 'PROPERTY':
            return WrenchIcon
        case 'FIELD':
            return TextboxIcon
        case 'CONSTRUCTOR':
            return CubeOutlineIcon
        case 'ENUM':
            return NumericIcon
        case 'INTERFACE':
            return LinkIcon
        case 'FUNCTION':
            return FunctionIcon
        case 'VARIABLE':
            return CubeIcon
        case 'CONSTANT':
            return PiBoxIcon
        case 'STRING':
            return CodeStringIcon
        case 'NUMBER':
            return PoundIcon
        case 'BOOLEAN':
            return MatrixIcon
        case 'ARRAY':
            return CodeArrayIcon
        case 'OBJECT':
            return DrawingBoxIcon
        case 'KEY':
            return KeyIcon
        case 'NULL':
            return NullIcon
        case 'ENUMMEMBER':
            return NumericIcon
        case 'STRUCT':
            return PillarIcon
        case 'EVENT':
            return TimetableIcon
        case 'OPERATOR':
            return CodeNotEqualIcon
        case 'TYPEPARAMETER':
            return CubeIcon
        case 'UNKNOWN':
        default:
            return ShapeIcon
    }
}

export interface SymbolIconProps {
    kind: GQL.SymbolKind
    className?: string
}

/**
 * Renders an Icon for a given symbol kind
 */
export const SymbolIcon: React.StatelessComponent<SymbolIconProps> = ({ kind, className = '' }) => {
    const Icon = getSymbolIconComponent(kind)
    return (
        <Icon
            className={`symbol-icon symbol-icon--kind-${kebabCase(kind)} ${className}`}
            data-tooltip={kind.toLowerCase()}
        />
    )
}
