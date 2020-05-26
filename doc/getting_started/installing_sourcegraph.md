# Installing Sourcegraph

Once Sourcegraph is deployed and running, the installation process takes only a few minutes. This screencast shows the typical steps involved in installing and configuring a private, self-hosted Sourcegraph instance, ready for sharing internally.

It uses GitHub as the code host, but the steps are largely the same for other code hosts, the only major difference being that GitLab and Bitbucket Server provide native code intelligence integrations.

<div style="padding:56.25% 0 0 0;position:relative;">
    <iframe src="https://www.youtube.com/embed/iVTroSw9dhQ" style="position:absolute;top:0;left:0;width:100%;height:100%;" frameborder="0" webkitallowfullscreen="" mozallowfullscreen="" allowfullscreen=""></iframe>
</div>

---

Below is a similar configuration example to the one used in the screencast.

```javascript
// Note that Sourcegraph configuration supports JSONC (JSON with comments).
{
  "url": "https://github.com",
  "token": "<token>",

  "orgs": [
    "sourcegraph"
  ],
  
  // Specific include list, e.g, open source code used internally
  "repos": [
    "hashicorp/terraform",
    "terraform-providers/terraform-provider-aws",
    "jenkinsci/jenkins",
    "facebook/react",
    "twbs/bootstrap",
    "django/django"
  ],

  // Filter out archived and forked repos from repositories included by "orgs"
  "exclude": [
    {
      "archived": true
    },
    {
      "forks": true
    }
  ]
}
```

[**» Next: Connecting your code host**](connecting_your_code_host.md)
