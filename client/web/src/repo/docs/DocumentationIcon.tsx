import classNames from 'classnames'
import { upperFirst } from 'lodash'
import { MdiReactIconComponentType } from 'mdi-react'
import AccountMultipleIcon from 'mdi-react/AccountMultipleIcon'
import ClockFastIcon from 'mdi-react/ClockFastIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import CodeArrayIcon from 'mdi-react/CodeArrayIcon'
import CodeBracesIcon from 'mdi-react/CodeBracesIcon'
import CodeNotEqualIcon from 'mdi-react/CodeNotEqualIcon'
import CodeStringIcon from 'mdi-react/CodeStringIcon'
import CubeIcon from 'mdi-react/CubeIcon'
import CubeOutlineIcon from 'mdi-react/CubeOutlineIcon'
import DrawingBoxIcon from 'mdi-react/DrawingBoxIcon'
import EyeOffIcon from 'mdi-react/EyeOffIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import FunctionIcon from 'mdi-react/FunctionIcon'
import GavelIcon from 'mdi-react/GavelIcon'
import KeyIcon from 'mdi-react/KeyIcon'
import LinkIcon from 'mdi-react/LinkIcon'
import MatrixIcon from 'mdi-react/MatrixIcon'
import NullIcon from 'mdi-react/NullIcon'
import NumericIcon from 'mdi-react/NumericIcon'
import PackageIcon from 'mdi-react/PackageIcon'
import PiBoxIcon from 'mdi-react/PiBoxIcon'
import PillarIcon from 'mdi-react/PillarIcon'
import PoundIcon from 'mdi-react/PoundIcon'
import SchoolIcon from 'mdi-react/SchoolIcon'
import ShapeIcon from 'mdi-react/ShapeIcon'
import SitemapIcon from 'mdi-react/SitemapIcon'
import TestTubeIcon from 'mdi-react/TestTubeIcon'
import TextBoxIcon from 'mdi-react/TextBoxIcon'
import TimetableIcon from 'mdi-react/TimetableIcon'
import WebIcon from 'mdi-react/WebIcon'
import WrenchIcon from 'mdi-react/WrenchIcon'
import * as React from 'react'

import styles from './DocumentationIcon.module.scss'
import { Tag } from './graphql'

/**
 * Returns the icon component for a given documentation tag
 */
const getDocumentationIconComponent = (tag: Tag): MdiReactIconComponentType => {
    switch (tag) {
        case 'private':
            return EyeOffIcon
        case 'deprecated':
            return CloseIcon
        case 'test':
            return TestTubeIcon
        case 'benchmark':
            return ClockFastIcon
        case 'example':
            return SchoolIcon
        case 'license':
            return GavelIcon
        case 'owner':
            return AccountMultipleIcon

        // Tags derived from SymbolKind
        case 'file':
            return FileDocumentIcon
        case 'module':
            return CodeBracesIcon
        case 'namespace':
            return WebIcon
        case 'package':
            return PackageIcon
        case 'class':
            return SitemapIcon
        case 'method':
            return CubeOutlineIcon
        case 'property':
            return WrenchIcon
        case 'field':
            return TextBoxIcon
        case 'constructor':
            return CubeOutlineIcon
        case 'enum':
            return NumericIcon
        case 'interface':
            return LinkIcon
        case 'function':
            return FunctionIcon
        case 'variable':
            return CubeIcon
        case 'constant':
            return PiBoxIcon
        case 'string':
            return CodeStringIcon
        case 'number':
            return PoundIcon
        case 'boolean':
            return MatrixIcon
        case 'array':
            return CodeArrayIcon
        case 'object':
            return DrawingBoxIcon
        case 'key':
            return KeyIcon
        case 'null':
            return NullIcon
        case 'enumNumber':
            return NumericIcon
        case 'struct':
            return PillarIcon
        case 'event':
            return TimetableIcon
        case 'operator':
            return CodeNotEqualIcon
        case 'typeParameter':
            return CubeIcon
        default:
            return ShapeIcon
    }
}

interface Props {
    tag: Tag
    className?: string
}

function getDocumentationIconClassName(tag: Tag): string | undefined {
    return styles[`documentationIconTag${upperFirst(tag)}` as keyof typeof styles]
}

/**
 * Renders an Icon for a given documentation tag
 */
export const DocumentationIcon: React.FunctionComponent<Props> = ({ tag, className = '' }) => {
    const Icon = getDocumentationIconComponent(tag)
    return (
        <Icon
            className={classNames(styles.documentationIcon, getDocumentationIconClassName(tag), className)}
            data-tooltip={tag}
        />
    )
}
