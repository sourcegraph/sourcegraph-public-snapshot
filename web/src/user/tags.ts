import * as GQL from '../backend/graphqlschema'

const PLATFORM_TAG = 'platform'

export function platformEnabled(user: Pick<GQL.IUser, 'tags'>): boolean {
    return !!window.context.platformEnabled && hasTag(user, PLATFORM_TAG)
}

function hasTag(user: Pick<GQL.IUser, 'tags'>, tag: string): boolean {
    return user.tags.includes(tag)
}
