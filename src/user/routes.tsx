import { userAreaRoutes } from '@sourcegraph/webapp/dist/user/area/routes'
import { UserAreaRoute } from '@sourcegraph/webapp/dist/user/area/UserArea'

export const enterpriseUserAreaRoutes: ReadonlyArray<UserAreaRoute> = [...userAreaRoutes]
