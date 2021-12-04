import { CatalogEntityRelationType } from '@sourcegraph/shared/src/graphql-operations'

const CATALOG_RELATION_TYPE_DISPLAY_NAMES: Record<CatalogEntityRelationType, string> = {
    DEPENDS_ON: 'Depends on',
    DEPENDENCY_OF: 'Dependency of',
    HAS_PART: 'Has part',
    PART_OF: 'Part of',
}

export function catalogRelationTypeDisplayName(edgeType: CatalogEntityRelationType): string | undefined {
    return CATALOG_RELATION_TYPE_DISPLAY_NAMES[edgeType]
}
