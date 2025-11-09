package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"

	"github.com/tzrikka/xdg"
)

const (
	fileOption = "file"

	DataDirName  = "thrippy"
	DataFileName = "secrets.toml"
	DataFilePerm = 0o600
)

type fileProvider struct {
	path string
	mu   sync.RWMutex
}

func newFileProvider() (Manager, error) {
	return &fileProvider{path: dataFile()}, nil
}

func (p *fileProvider) Set(_ context.Context, key, value string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	store, err := p.readTOMLFile()
	if err != nil {
		return err
	}

	store[key] = value
	if err := p.writeTOMLFile(store); err != nil {
		return err
	}
	return nil
}

func (p *fileProvider) Get(_ context.Context, key string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	store, err := p.readTOMLFile()
	if err != nil {
		return "", err
	}

	v, ok := store[key]
	if !ok {
		return "", nil
	}
	return v, nil
}

func (p *fileProvider) Delete(_ context.Context, key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	store, err := p.readTOMLFile()
	if err != nil {
		return err
	}

	delete(store, key)
	if err := p.writeTOMLFile(store); err != nil {
		return err
	}
	return nil
}

// dataFile returns the path to the app's data file.
// It also creates an empty file if it doesn't already exist.
func dataFile() string {
	path, err := xdg.CreateFile(xdg.DataHome, DataDirName, DataFileName)
	if err != nil {
		log.Fatal().Err(err).Caller().Msg("failed to create data file")
	}
	return path
}

func (p *fileProvider) readTOMLFile() (map[string]string, error) {
	var tomlStore map[string]any
	if _, err := toml.DecodeFile(p.path, &tomlStore); err != nil {
		return nil, err
	}

	flatStore := map[string]string{}
	for prefix, v1 := range tomlStore {
		m1, ok := v1.(map[string]any)
		if !ok {
			continue
		}
		for namespace, v2 := range m1 {
			m2, ok := v2.(map[string]any)
			if !ok {
				continue
			}
			for id, v3 := range m2 {
				m3, ok := v3.(map[string]any)
				if !ok {
					continue
				}
				for field, v4 := range m3 {
					s, ok := v4.(string)
					if !ok {
						continue
					}
					k := fmt.Sprintf("%s/%s/%s/%s", prefix, namespace, id, field)
					flatStore[k] = s
				}
			}
		}
	}

	return flatStore, nil
}

func (p *fileProvider) writeTOMLFile(store map[string]string) error {
	f, err := os.OpenFile(p.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, DataFilePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	// Ensure file permissions are set correctly even if the file already exists.
	if err := os.Chmod(p.path, DataFilePerm); err != nil {
		return err
	}

	tomlStore := map[string]any{}
	for k, v := range store {
		tomlKeys := strings.Split(k, "/")
		m := tomlStore
		for i := range 3 {
			subMap, ok := m[tomlKeys[i]]
			if !ok {
				m[tomlKeys[i]] = map[string]any{}
				subMap = m[tomlKeys[i]]
			}
			m = subMap.(map[string]any)
		}
		m[tomlKeys[3]] = v
	}

	if err := toml.NewEncoder(f).Encode(tomlStore); err != nil {
		return err
	}
	return nil
}
