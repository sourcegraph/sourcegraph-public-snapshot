# Site admin configuration for Batch Changes

#### Setup Batch Changes 

<ol>
  <li>
    Using Batch Changes requires a <a href="../../../admin/external_service">code host connection</a> to a supported code host (currently GitHub, Bitbucket Server / Bitbucket Data Center, GitLab, and Bitbucket Cloud).
  </li>
  <li>
    (Optional) <a href="../../../admin/repo/permissions">Configure repository permissions</a>, which Batch Changes will respect.
  </li>
  <li>
    <a href="configuring_credentials">Configure credentials</a>.
  </li>
  <li>
    Setup webhooks to make sure changesets sync fast. See <a href="../references/requirements#batch-changes-effect-on-code-host-rate-limits">Batch Changes effect on codehost rate limits</a>.
    <ul>
      <li>
        <a href="../../admin/external_service/github#webhooks">GitHub</a>
      </li>
      <li>
        <a href="../../admin/external_service/bitbucket_server#webhooks">Bitbucket Server / Bitbucket Data Center</a>
      </li>
      <li>
        <a href="../../admin/external_service/gitlab#webhooks">GitLab</a>
      </li>
      <li>
        <a href="../../admin/external_service/bitbucket_cloud#webhooks">Bitbucket Cloud</a>
      </li>
    </ul>
    <aside class="note">
      NOTE: Incoming webhooks can be viewed in <strong>Site Admin &gt; Batch Changes &gt; Incoming webhooks</strong>. Webhook logging can be configured through the <a href="../../admin/config/batch_changes#incoming-webhooks">incoming webhooks site configuration</a>.
    </aside>
  </li>
  <li>
    Configure any desired optional features, such as:
    <ul>
      <li>
        <a href="../../../admin/config/batch_changes#rollout-windows">Rollout windows</a>, which control the rate at which Batch Changes will publish changesets on code hosts.
      </li>
      <li>
        <a href="../../../admin/config/batch_changes#forks">Forks</a>, which push branches created by Batch Changes onto forks of the upstream repository instead than the repository itself.
    </ul>
  </li>
</ol>


#### Disable Batch Changes
- [Disable Batch Changes](../explanations/permissions_in_batch_changes.md#disabling-batch-changes).
- [Disable Batch Changes for non-site-admin users](../explanations/permissions_in_batch_changes.md#disabling-batch-changes-for-non-site-admin-users).
