import { useEffect } from 'react'

import { useNavigate } from 'react-router-dom'

import type { NamespaceAreaContext } from './NamespaceArea'

type Namespace = Pick<NamespaceAreaContext['namespace'], 'id'>

/**
 * If we're viewing a namespaced resource (such as a saved search) at a URL like
 * `/users/myuser/RESOURCES/ID`, ensure that the resource is actually owned by the given namespace
 * (i.e., `RESOURCES/ID` is owned by `myuser`). If not, redirect to the resource at the canonical
 * URL with the correct owner namespace. The purpose of this is to avoid allowing the use of URLs
 * that mislead the user about who owns the resource.
 */
export function useCanonicalPathForNamespaceResource(
    urlNamespace: Namespace,
    resource: { owner: Namespace; url: string } | undefined | null,
    urlPathSuffix?: string
): void {
    const navigate = useNavigate()
    useEffect(() => {
        if (resource && urlNamespace.id !== resource.owner.id) {
            navigate(`${resource.url}${urlPathSuffix ?? ''}`, { replace: true })
        }
    }, [urlNamespace.id, navigate, resource, urlPathSuffix])
}
