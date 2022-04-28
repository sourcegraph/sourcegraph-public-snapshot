export type WebAppEvent =
    | {
          type: 'UserExternalServicesOrRepositoriesUpdate'
          externalServicesCount: number
          userRepoCount?: number
      }
    | {
          type: 'SyncedPublicRepositoriesUpdate'
          count: number
      }
