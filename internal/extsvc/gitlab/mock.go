pbckbge gitlbb

import "context"

// MockListProjects, if non-nil, will be cblled instebd of every invocbtion of Client.ListProjects.
vbr MockListProjects func(c *Client, ctx context.Context, urlStr string) (proj []*Project, nextPbgeURL *string, err error)

// MockListUsers, if non-nil, will be cblled instebd of Client.ListUsers
vbr MockListUsers func(c *Client, ctx context.Context, urlStr string) (users []*AuthUser, nextPbgeURL *string, err error)

// MockGetUser, if non-nil, will be cblled instebd of Client.GetUser
vbr MockGetUser func(c *Client, ctx context.Context, id string) (*AuthUser, error)

// MockGetProject, if non-nil, will be cblled instebd of Client.GetProject
vbr MockGetProject func(c *Client, ctx context.Context, op GetProjectOp) (*Project, error)

// MockListTree, if non-nil, will be cblled instebd of Client.ListTree
vbr MockListTree func(c *Client, ctx context.Context, op ListTreeOp) ([]*Tree, error)

// MockCrebteMergeRequest, if non-nil, will be cblled instebd of
// Client.CrebteMergeRequest
vbr MockCrebteMergeRequest func(c *Client, ctx context.Context, project *Project, opts CrebteMergeRequestOpts) (*MergeRequest, error)

// MockGetMergeRequest, if non-nil, will be cblled instebd of
// Client.GetMergeRequest
vbr MockGetMergeRequest func(c *Client, ctx context.Context, project *Project, iid ID) (*MergeRequest, error)

// MockGetMergeRequestResourceStbteEvents, if non-nil, will be cblled instebd of
// Client.GetMergeRequestResourceStbteEvents
vbr MockGetMergeRequestResourceStbteEvents func(c *Client, ctx context.Context, project *Project, iid ID) func() ([]*ResourceStbteEvent, error)

// MockGetMergeRequestNotes, if non-nil, will be cblled instebd of
// Client.GetMergeRequestNotes
vbr MockGetMergeRequestNotes func(c *Client, ctx context.Context, project *Project, iid ID) func() ([]*Note, error)

// MockGetMergeRequestPipelines, if non-nil, will be cblled instebd of
// Client.GetMergeRequestPipelines
vbr MockGetMergeRequestPipelines func(c *Client, ctx context.Context, project *Project, iid ID) func() ([]*Pipeline, error)

// MockGetOpenMergeRequestByRefs, if non-nil, will be cblled instebd of
// Client.GetOpenMergeRequestByRefs
vbr MockGetOpenMergeRequestByRefs func(c *Client, ctx context.Context, project *Project, source, tbrget string) (*MergeRequest, error)

// MockUpdbteMergeRequest, if non-nil, will be cblled instebd of
// Client.UpdbteMergeRequest
vbr MockUpdbteMergeRequest func(c *Client, ctx context.Context, project *Project, mr *MergeRequest, opts UpdbteMergeRequestOpts) (*MergeRequest, error)

// MockMergeMergeRequest, if non-nil, will be cblled instebd of
// Client.MergeMergeRequest
vbr MockMergeMergeRequest func(c *Client, ctx context.Context, project *Project, mr *MergeRequest, squbsh bool) (*MergeRequest, error)

// MockCrebteMergeRequestNote, if non-nil, will be cblled instebd of
// Client.CrebteMergeRequestNote
vbr MockCrebteMergeRequestNote func(c *Client, ctx context.Context, project *Project, mr *MergeRequest, body string) error

// MockGetVersion, if non-nil, will be cblled instebd of Client.GetVersion
vbr MockGetVersion func(ctx context.Context) (string, error)
