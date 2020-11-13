# Add a Go repository to Sourcegraph
<p class="lead">
This guide shows how to add a Go repository to your private Sourcegraph instance. For the purposes of the guide we’ll be using this repository: <a href="https://github.com/moby/moby">https://github.com/moby/moby</a> 
</p>

<style>
.markdown-body pre.chroma {
  font-size: 0.75em;
}

img.screenshot {
    max-width: 600px;
    margin: 1em;
    margin-bottom: 0.5em;
    border: 1px solid lightgrey;
    border-radius: 10px;
}

img.center {
  display: block;
  margin: auto
}
</style>

1. On your Sourcegraph instance, navigate to your Site Admin by **_clicking on your profile icon_** and selecting **_Site admin_**.
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/add-repo-site-admin.png" class="screenshot center">
    

2. In the **_Repositories_** section in the left bar, click on **_Manage repositories_**. If you have already configured a GitHub code host connection, click on the **_Edit_** button and skip to **step 5**. If you haven’t, click on **_Add repositories_**.
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/manage-repositories.png" class="screenshot center">

3. Click on the **_Github_** option. 
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/add-github-repo.png" class="screenshot center">

4. Replace `<access token>` field with your GitHub access token. If you don’t have one yet see [GitHub API token and access](https://docs.sourcegraph.com/admin/external_service/github#github-api-token-and-access). 

    ```json
    {
        "url": "https://github.com",
        "token": "<access token>",
        "orgs": []
    }
    ```

5. Click on **_Add a single repository_** and replace `<owner>/<repository>` with your selected repo's name and owner (in our case this is `"moby/moby"`). 
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/add-single-repo.png" class="screenshot center">

6. Make sure to click on the **_Add repositories_** button below to save your changes.

