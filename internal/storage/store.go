package storage

type StoreOptions struct {
	RootPath string
}

func DefaultStoreOptions() StoreOptions {
	return StoreOptions{
		RootPath: "/var/lib/cbt",
	}
}

type Store interface {
	RootPath() string
}

type store struct {
	rootPath string
}

func New(options StoreOptions) (Store, error) {
	return &store{
		rootPath: options.RootPath,
	}, nil
}

func (s *store) RootPath() string {
	return s.rootPath
}
