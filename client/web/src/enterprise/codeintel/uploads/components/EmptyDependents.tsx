import React from 'react'


import { Text, Icon } from '@sourcegraph/wildcard'
import { mdiMapSearch } from "@mdi/js";

export const EmptyDependents: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Text className="text-muted text-center w-100 mb-0 mt-1">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        This upload has no dependents.
    </Text>
)
