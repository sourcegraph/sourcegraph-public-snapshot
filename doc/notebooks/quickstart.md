<!-- TODO: Get all images uploaded to GCP -->
# Quickstart for Notebooks
Create your first Notebook in minutes.
# Introduction
In this guide, you'll create your first Notebook
# The Notebooks interface
The Notebooks interface will be familiar if you've ever used Jupyter or iPython notebooks before.
![A new notebook](https://firebasestorage.googleapis.com/v0/b/clover-f2488.appspot.com/o/document-images%2Fcf96cf05bd95f5e6ebe2cf9d1bf3c89a00ab4048?alt=media&token=37cbb86f-be98-442b-9166-3b9a7918a36e)
### Run all blocks
The Run all blocks button allows you to render all the blocks in the Notebook with one click. You'll want to get in the habit of using this button every time you start reading a Notebook.
![The Run all blocks button](https://firebasestorage.googleapis.com/v0/b/clover-f2488.appspot.com/o/document-images%2Fbcf6bb469a06aec007ba95e471f6806d7764062a?alt=media&token=a7a277a2-0545-49fb-b155-f4c7a53d7fa3)
### Autosave
Notebooks are automatically saved so you'll never lose your work.
![null](https://firebasestorage.googleapis.com/v0/b/clover-f2488.appspot.com/o/document-images%2F6907d7bbb1a22c1a3b22b3cb9090e34e66e44144?alt=media&token=a66dbba7-be2e-4666-96c3-09a5bf8fb86c)
### Supported block types
Like Jupyter notebooks, Sourcegraph Notebooks support Markdown. They also support two special block types: Query and Code.

#### Query blocks
Query blocks allow you to enter any valid Sourcegraph query and view the results right inside the notebook. The query block input supports the same typeahead functionality as the main Sourcegraph search bar.
![](https://firebasestorage.googleapis.com/v0/b/clover-f2488.appspot.com/o/document-images%2F9b7a69e6cd03ad934a53933bcf5b95f3ec484f78?alt=media&token=b55ef013-4c18-45c7-9736-431cf932bbc3)

#### Markdown blocks
Markdown blocks support all the Markdown features you'd expect, including rendering images and syntax highlighting.

#### Code blocks
Code blocks are another special block type that allow you show a file range from any file on Sourcegraph. You can use the interface to fill in the range you want, or just select a code block and paste in a valid Sourcegraph URL to have the inputs auto-filled.

![CleanShot 2022-02-05 at 17.50.23](https://cleanshot-cloud-fra.accelerator.net/media/21250/4Mz4hkFStI53vv7BKYbaqIc64Rr7aYQ6yDgx2DTo.mp4;frame(1)/resize(1200,630,fit)/draw(image(b64-L2ludGVybmFsL3BsYXkucG5n),position:center).jpeg)

# Create a new Notebook
- Navigate to [https://sourcegraph.com/notebooks](https://sourcegraph.com/notebooks) and click "Create notebook".
- Click on the name to edit and rename the Notebook "React components"
- Create a Markdown block and enter the text below.

```
# React class components
React class components are no longer the preferred style of component creation. Today, functional components are the standard. Below are some examples of class components in the Sourcegraph codebase.

```
- Click the render button, or press Cmd+Return to render the block
- Next, create a query block and enter the following query to find some examples of class components in the Sourcegraph codebase.


```
repo:^github\.com/sourcegraph/sourcegraph$ lang:TypeScript extends React
```

- Render the block with Cmd+Return
- Next, let's find some recent commits that include class components.
- Click the "duplicate" button to duplicate the block and modify the query to return commits that added our search phrase.


```
repo:^github\.com/sourcegraph/sourcegraph$ lang:TypeScript extends React type:diff select:commit.diff.added
```

- Your Notebook should now look like [this](https://sourcegraph.com/notebooks/Tm90ZWJvb2s6MTE4).
- **Optional:** share you notebook by clicking "Private" at the top right of the screen and select "Public" to share your Notebook with the community!

# Next steps
Congratulations on creating your first Notebook! Next, get some inspiration from the Notebooks created by Sourcegraph team members and the wider community by [exploring public Notebooks](https://sourcegraph.com/notebooks?tab=explore).
