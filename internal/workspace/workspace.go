package workspace

import (
	"os"
	"path/filepath"
)

type Workspace struct {
	Dir  string
	keep bool
}

func New(baseDir string, prefix string, keep bool) (*Workspace, error) {
	dir, err := os.MkdirTemp(baseDir, prefix)
	if err != nil {
		return nil, err
	}
	return &Workspace{
		Dir:  dir,
		keep: keep,
	}, nil
}

func (w *Workspace) Path(name string) string {
	return filepath.Join(w.Dir, name)
}

func (w *Workspace) Cleanup() error {
	// keep the directory in /tmp if true
	if w.keep {
		return nil
	}
	return os.RemoveAll(w.Dir)
}
