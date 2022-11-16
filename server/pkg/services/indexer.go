package services

func (i *indexService) Indexer() error {
	i.storage.Indexer()
	return nil
}
