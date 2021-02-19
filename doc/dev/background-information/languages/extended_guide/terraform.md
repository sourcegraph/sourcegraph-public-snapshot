# Playbooks for Terraform

## Statefile surgery

_Note: these are provided as internal reference documents. You should rarely need to use this and please reach out to the distribution team or ask in the #dev-ops channel before performing these steps._

To debug some more involved terraform problems you may need to directly modify the state file.

1. Use `TF_DEBUG` environment variable with `WARN`, `DEBUG` logs enabled to locate a possible source of the issue.

1. Use `terraform state pull` to examine the statefile locally.

1. If possible, use `terraform state rm`/`terraform state mv X Y` to make the necessary state modifications.
   Otherwise, manually edit the state file and run `terraform state push` to update the remote state.

    1. In cases where a corrupted state has been pushed, you can pull a previous object version with [gsutil](https://cloud.google.com/storage/docs/using-object-versioning)

1. File bugs as appropriate, statefile modification should typically not be necessary
