// hasTag returns whether the user or org has the given tag.
export function hasTag(arg: GQL.IUser | GQL.IOrg, tag: string): boolean {
    return arg.tags && (arg.tags as { name: string }[]).some(tag2 => tag2.name === tag)
}

// hasTagRecursive returns whether the user or org has the given tag.
// For a user, it returns whether the user or any of their orgs have
// the tag.
export function hasTagRecursive(arg: GQL.IUser | GQL.IOrg | null, tag: string): boolean {
    if (!arg) {
        return false
    }
    return hasTag(arg, tag) || (isUser(arg) && arg.orgs && arg.orgs.some(org => hasTag(org, tag)))
}

function isUser(arg: GQL.IUser | GQL.IOrg): arg is GQL.IUser {
    return !!(arg as any).orgs
}
