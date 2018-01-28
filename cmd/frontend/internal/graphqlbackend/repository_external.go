package graphqlbackend

func (r *repositoryResolver) ExternalRepository() *externalRepositoryResolver {
	if r.repo.ExternalRepo == nil {
		return nil
	}
	return &externalRepositoryResolver{repository: r}
}

type externalRepositoryResolver struct {
	repository *repositoryResolver
}

func (r *externalRepositoryResolver) ID() string { return r.repository.repo.ExternalRepo.ID }
func (r *externalRepositoryResolver) ServiceType() string {
	return r.repository.repo.ExternalRepo.ServiceType
}
func (r *externalRepositoryResolver) ServiceID() string {
	return r.repository.repo.ExternalRepo.ServiceID
}
