package tst

type ScenarioResource struct {
	name string
	id   string
	key  string
}

func NewScenarioResource(name string) *ScenarioResource {
	id := id()
	key := joinID(name, "-", id, 39)
	return &ScenarioResource{
		name: name,
		id:   id,
		key:  key,
	}

}

func (s *ScenarioResource) ID() string {
	return s.id
}

func (s *ScenarioResource) Name() string {
	return s.name
}

func (s *ScenarioResource) Key() string {
	return s.key
}
