# Cody App troubleshooting guide

### VS Code extension isn't indexing repositories correctly
If you've confirmed that your repositories have been added to your app and embeddings have been generated successfully (check Settings > Advanced settings > Embedding jobs), the issue may be that your repositories don't have a git remote. If that's the case, try updating your extension settings (Settings > Settings > Extensions > Cody) to set `Cody: Codebase` to your path. For example, if the path on your local repositories pages is `User/myname/projects/my-repo`, you'd set `Cody: Codebase` to `my-repo`. 

 If you're having trouble, please file an issue on the [issue tracker](https://github.com/sourcegraph/app/issues) or ask in our [Discord](https://discord.gg/s2qDtYGnAE).