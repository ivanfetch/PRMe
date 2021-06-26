package prme

type client struct{}

func New(token string) (*client, error) {
	return &client{}, nil
}

func (c client) GetRepoID(repo string) (int, error) {
	return 0, nil
}
