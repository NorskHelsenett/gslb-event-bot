package persistence

import "iter"

type Store[T any] interface {
	Save(key string, data T) error
	Load(key string) (T, error)
	LoadAll() (iter.Seq[T], func() error)
	Delete(key string) error
	Close() error
}

type MigrateFunc func(map[string]any) (map[string]any, error)

func Chain(fns ...MigrateFunc) MigrateFunc {
	return func(raw map[string]any) (map[string]any, error) {
		var err error
		for _, fn := range fns {
			raw, err = fn(raw)
			if err != nil {
				return nil, err
			}
		}
		return raw, nil
	}
}
