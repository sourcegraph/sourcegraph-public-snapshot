# Access control

Controlling who can have access to a Sourcegraph instance can be set via filters available in the [site configuration](../config/site_config.md).

Depending on the type of authentication you decided to use (see here more details about [SSO options](../auth/index.md)), you can choose the following ways to restrict sign-in and sign-up for new or existing users.

## GitHub filters

**allowSignup**

  Allows new users to sing up to Sourcegraph via GitHub authentication (by clicking the "Continue with GitHub" button).
    
  If set to `false` or not set, sign-up will be blocked. In this case, new users will only be able to sign in after an admin creates their account on Sourcegraph.
  
  The new user email, during their account creation, should match one of their GitHub verified emails.

  ````
    {
      "type": "github",
      ...
      "allowSignup": false
    }
  ````

**allowOrgs**

  Restricts logins to members of the listed organizations. If empty or unset, no restriction will be applied.

  If combined with `"allowSignup": true`, only membrers of the the allowed orgs can create their accounts in Sourcegraph via GitHub authentitcation.
  When `"allowSignup": false`, an admin should first create the user account so the user can then sign in with GitHub if they belong to the allowed orgs.

  ````
    {
      "type": "github",
      ...
      "allowOrgs": ["org1", "org2"] 
    },
  ````
 

**allowOrgsMap**

  Restricts sign-ups and new logins to members of the listed teams or subteams that need to be mapped to their parent organization.

  If combined with `"allowSignup": true`, only membrers of the the allowed teams can create their accounts in Sourcegraph via GitHub authentitcation.
  When combined with `"allowSignup": false`, an admim should first create the user account so that the user can login with GitHub.

  Note that subteams inheritance is not supported — the name of child teams (subteams) should be informed so their members can be granted access to Sourcegraph.
  
  If empty or unset, no restrictions will be applied.

  ````
    {
       "type": "github"
      ... 
      "allowOrgsMap": {
        "org1": [
          "team1", "subteam1"
        ],
        "org2": [
          "subteam2"
        ]
      }
    }
  ````


## GitLab filters

**allowSignup**

  Allows new users to create their accounts in Sourcegraph via GitLab authentication (by clicking the "Continue with GitLab" button).
  
  When `false`, sign-up will be blocked. In this case, new users can only sign in after an admin creates their account on Sourcegraph. The user account email should match their GitLab email.

  If combined with `"allowSignup": true`, only membrers of the the allowed groups/subgroups can create their accounts in Sourcegraph via GitLab authentitcation.
  When combined with `"allowSignup": false`, an admim should first create the user account so that the user can login with GitLab.

  *If not set, unlinke with GitHub, it allowSignup defaults to `true`*.

  ````
    {
      "type": "gitlab",
      ...
      "allowSignup": false
    }
  ````

**allowGroups**

  Restricts new logins to members of the listed groups or subgroups. 
  
  Instead of informing the groups or subgroups names, use their full path that can be copied from the URL.

  For a parent group, the full path will be simple as "group", but for nested groups it can look like “group/subgroup/subsubgroup”.

  When empty or unset, no restrictions will be applied.


  If combined with `"allowSignup": false`, an admim should first create the user account so that the user can sign in with GitLab.

  If combined with `"allowSignup": true`, only membrers of the the allowed groups or subgroups can create their accounts in Sourcegraph via GitLab authentitcation.

  ````
    {
      "type": "gitlab",
      ...
      "allowSignup": true,
      "allowGroups": [
        "group/subgroup/subsubgroup"
      ]
    }
  ````

## SAML filters 

**allowSignup**

  It works the same way as in GitHub or Github, allowing new users to creating their accounts via SAML authentication, or blocking the sign-up when set to `false`.

  If false, users signing in via SAML must have an existing Sourcegraph account, which will be linked to their SAML identity after the sign-in.

  ````
    {
      type: "saml",
      ...
      "allowSignup": true
    }
  ````

**allowGroups**

  Restricts login to members of the allowed SAML groups.

  When not configured or set to`true`, sign-in will be allowed.
  If the list of allowed groups is empty, sign-in is not allowed.

  The `groupAttributesName` is optional and will default to "groups" when not informed.

  ````
    {
      type: "saml",
      ...
      "allowGroups": ["sourcegraph"]
      "groupAttributesName": "mySAMLgroup"
    }
  ````

## OpenID filter

**allowSignup**

  It allows new users to creating their accounts via OpenID, or blocks the sign-up when set to `false`.

  ````
    {
      type: "openidconnect",
      ...
      "allowSignup": false
    }
  ````

## Builtin filter

**allowSignup**

  If `true`, users will see the sign-up link under the login form and will have access to the sign-up page, where they can create their accounts without restriction.

  If you choose to block sign-up in another auth provider, make sure this builtin filter is removed or set to `false`. Otherwise, users will have a way to bypass the restriction.

  During the initial setup, the builtin sign-up will be available for the first user so they can create an account and become admin.

  ```
    "auth.providers": [
      {
        "type": "builtin",
        "allowSignup": false
      }
    ]
  ````
