# Access control

Controlling who can have access to a Sourcegraph instance can be done via the site configuration. 

Depending on the type of authentication you decided to use (see more details here about our SSO options), you can choose the following ways to restrict sign-in and sign-up for new or existing users.

## GitHub filters

**AllowSignup**

  Allows new users to create their accounts in Sourcegraph via GitHub authentication (by clicking the "Continue with GitHub" button)
    
  When set to `false` or not set, signup will be blocked. In this case, new users will only be able to sign in after an admin creates their account on Sourcegraph. 
  
  Note that the new user email, when their account is created by the admin, should match one of their GitHub verified emails.


  ````
    {
      ...
      "allowSignup": false,
    }
  ````


### AllowOrgs

  Restricts signups and logins to members of the listed organizations. If empty or unset, no restriction will be applied.

  ````
     {
      ...
      "allowOrgs": ["org1", "org2"] 
    },
  ````
 

### AllowOrgsMap

  Restricts new logins to members of the listed teams or subteams that need to be mapped to their parent organization. 
  
  Note that subteams inheritance is not supported — the name of teams and subteams should be explicated informed so their members can be granted access to Sourcegraph. If this filter is empty or unset, no restrictions will be applied.

  ````
    {
      ... 
      "allowOrgsMap": {
        "org1": [
          "team1", "team2"
        ],
        "org2": [
          "team3"
        ]
      }
    }
  ````


## GitLab filters

**AllowSignup**

  Allows new users to create their accounts in Sourcegraph via GitLab authentication (by clicking the "Continue with GitLab" button)
  
  When set to `false` or not set, signup will be blocked. In this case, new users will only be able to sign in after an admin creates an account for them on Sourcegraph, informing their GitLab email.

  The new user email, when their account is created, should match their GitLab email.

    ````
    ````

**AllowGroups**

  Restricts new logins to members of the listed groups or subgroups. 
  
  Instead of informing the groups or subgroups names, use the full path that can be copied from the URL. It looks like: “group/subgroup/subsubgroup”

  When empty or unset, no restrictions will be applied.

    ````
    ````

## SAML filters 


  ````
  ````
## Builtin filter

   **AllowSignup**

  If set to true, users will see the signup link under the login form and will have access to the signup page, where they can create their new accounts without restriction.

  if you choose to block signup in another auth provider, don’t forget to remove this filter or set it to false, otherwise, users can bypass the restriction. 

  During the initial setup, the builtin signup will be available for the first user that, after creating their account, will become admin.  

  ```
    "auth.providers": [
      {
        "type": "builtin",
        "allowSignup": false
      }
    ]
  ````
