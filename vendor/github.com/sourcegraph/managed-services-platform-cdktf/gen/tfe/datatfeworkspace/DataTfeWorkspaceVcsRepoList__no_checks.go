//go:build no_runtime_type_checking

package datatfeworkspace

// Building without runtime type checking enabled, so all the below just return nil

func (d *jsiiProxy_DataTfeWorkspaceVcsRepoList) validateGetParameters(index *float64) error {
	return nil
}

func (d *jsiiProxy_DataTfeWorkspaceVcsRepoList) validateResolveParameters(_context cdktf.IResolveContext) error {
	return nil
}

func (j *jsiiProxy_DataTfeWorkspaceVcsRepoList) validateSetTerraformAttributeParameters(val *string) error {
	return nil
}

func (j *jsiiProxy_DataTfeWorkspaceVcsRepoList) validateSetTerraformResourceParameters(val cdktf.IInterpolatingParent) error {
	return nil
}

func (j *jsiiProxy_DataTfeWorkspaceVcsRepoList) validateSetWrapsSetParameters(val *bool) error {
	return nil
}

func validateNewDataTfeWorkspaceVcsRepoListParameters(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, wrapsSet *bool) error {
	return nil
}

