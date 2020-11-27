# Enable LSIF in your global settings

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

<p class="lead">
Modify your Sourcegraph settings to enable precise code intelligence.
</p>

1. Navigate to your settings by clicking on your profile icon and selecting **_Settings_**.
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/enable-lsif-settings.png" class="screenshot">


2. Add `"codeIntel.lsif": true `to your settings and **_Save changes_** to enable LSIF.
    <img src="https://sourcegraphstatic.com/docs/images/code-intelligence/enable-lsif-save.png" class="screenshot">