package api

type IndexService interface {
	IndexMail() error
}

type indexService struct {
}

func NewIndexService() IndexService {
	return &indexService{}
}

func (s *indexService) IndexMail() error {
	return nil
}
