import React from 'react'

import { mdiArrowRight } from '@mdi/js'

import { Icon, Text } from '@sourcegraph/wildcard'

export const CloudCtaBanner: React.FunctionComponent<React.PropsWithChildren<{ children: React.ReactNode }>> = ({
    children,
}) => (
    <section className="my-3 p-2 d-flex justify-content-center bg-primary-4">
        <Icon className="mr-2 text-merged" size="md" aria-hidden={true} svgPath={mdiArrowRight} />

        <Text className="mb-0">{children}</Text>
    </section>
)
