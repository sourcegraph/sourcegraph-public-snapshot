# Add a GitHub repository to Sourcegraph
<p class="lead">
This guide shows how to add a GitHub repository to your Sourcegraph instance. For the purposes of the guide we’ll be using <a href="https://github.com/moby/moby">https://github.com/moby/moby</a> 
</p>

<style>
.markdown-body pre.chroma {
  font-size: 0.75em;
}

img.screenshot {
    display: block;
    margin: 1em auto;
    max-width: 600px;
    margin-bottom: 0.5em;
    border: 1px solid lightgrey;
    border-radius: 10px;
}
</style>

1. On your Sourcegraph instance, navigate to your Site Admin by **_clicking on your profile icon_** and selecting **_Site admin_**.
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/add-repo-site-admin.png" class="screenshot">
    

2. In the **_Repositories_** section in the left bar, click on **_Manage repositories_**. If you have already configured a GitHub code host connection, click on the **_Edit_** button and skip to **step 5**. If you haven’t, click on **_Add repositories_**.
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/manage-repositories.png" class="screenshot">

3. Click on the **_Github_** option. 
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/add-github-repo.png" class="screenshot">

4. Replace `<access token>` with your GitHub access token. If you don’t have one yet see [GitHub API token and access](https://docs.sourcegraph.com/admin/external_service/github#github-api-token-and-access). 

    ```json
    {
        "url": "https://github.com",
        "token": "<access token>",
        "orgs": []
    }
    ```

5. Click on **_Add a single repository_** and replace `<owner>/<repository>` with your selected repo's name and owner (in this case `"moby/moby"`). 
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/add-single-repo.png" class="screenshot">

6. Make sure to click on the **_Add repositories_** button below to save your changes.

