package blobs

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/blobs/replication"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const onlyForMe = 0700

const partialFileSuffix = ".part"

const filenameSeparator = "-"

const charactersInDirName = 2

type FilesystemStorage struct {
	path   string
	logger logging.Logger
}

func NewFilesystemStorage(path string, logger logging.Logger) (*FilesystemStorage, error) {
	s := &FilesystemStorage{
		path:   path,
		logger: logger,
	}

	if err := s.removeTemporaryFiles(); err != nil {
		return nil, errors.Wrap(err, "failed to remove old temporary files")
	}

	return s, nil
}

func (f FilesystemStorage) Store(id refs.Blob, r io.Reader) error {
	if err := f.ensureDirectoriesExist(); err != nil {
		return errors.Wrap(err, "error ensuring that directories exist")
	}

	hexRef := hex.EncodeToString(id.Bytes())

	pattern := fmt.Sprintf("%s%s%d%s*%s", hexRef, filenameSeparator, time.Now().Unix(), filenameSeparator, partialFileSuffix)

	tmpFile, err := os.CreateTemp(f.dirTemporary(), pattern)
	if err != nil {
		return errors.Wrap(err, "could not create a temporary file")
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	h := blobs.NewHasher()

	if _, err := io.Copy(io.MultiWriter(tmpFile, h), io.LimitReader(r, blobs.MaxBlobSize().InBytes())); err != nil {
		return errors.Wrap(err, "failed to copy contents to a temporary file")
	}

	if err := blobs.Verify(id, h); err != nil {
		return errors.Wrap(err, "failed to verify the file")
	}

	if err := tmpFile.Close(); err != nil {
		return errors.Wrap(err, "failed to close the temporary file")
	}

	if err := f.moveTemporaryFileToTargetFile(tmpFile, id); err != nil {
		return errors.Wrap(err, "failed to move the temporary file")
	}

	return nil
}

func (f FilesystemStorage) Create(r io.Reader) (refs.Blob, error) {
	if err := f.ensureDirectoriesExist(); err != nil {
		return refs.Blob{}, errors.Wrap(err, "error ensuring that directories exist")
	}

	pattern := fmt.Sprintf("%d%s*%s", time.Now().Unix(), filenameSeparator, partialFileSuffix)

	tmpFile, err := os.CreateTemp(f.dirTemporary(), pattern)
	if err != nil {
		return refs.Blob{}, errors.Wrap(err, "could not create a temporary file")
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	h := blobs.NewHasher()

	if _, err := io.Copy(io.MultiWriter(tmpFile, h), io.LimitReader(r, blobs.MaxBlobSize().InBytes())); err != nil {
		return refs.Blob{}, errors.Wrap(err, "failed to copy contents to a temporary file")
	}

	id, err := h.SumRef()
	if err != nil {
		return refs.Blob{}, errors.Wrap(err, "failed to calculate ref")
	}

	if err := tmpFile.Close(); err != nil {
		return refs.Blob{}, errors.Wrap(err, "failed to close the temporary file")
	}

	if err := f.moveTemporaryFileToTargetFile(tmpFile, id); err != nil {
		return refs.Blob{}, errors.Wrap(err, "failed to move the temporary file")
	}

	return id, nil
}

func (f FilesystemStorage) moveTemporaryFileToTargetFile(tmpFile *os.File, id refs.Blob) error {
	oldName := tmpFile.Name()
	newName := f.pathStorage(id)

	targetDir, _ := filepath.Split(newName)
	if err := os.MkdirAll(targetDir, onlyForMe); err != nil {
		return errors.Wrap(err, "error creating target directory")
	}

	if err := os.Rename(oldName, newName); err != nil {
		return errors.Wrap(err, "failed to rename the file")
	}

	return nil
}

func (f FilesystemStorage) Get(id refs.Blob) (io.ReadCloser, error) {
	name := f.pathStorage(id)
	return os.Open(name)
}

func (f FilesystemStorage) Size(id refs.Blob) (blobs.Size, error) {
	name := f.pathStorage(id)
	fi, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return blobs.Size{}, replication.ErrBlobNotFound
		}
		return blobs.Size{}, errors.Wrap(err, "stat failed")
	}
	return blobs.NewSize(fi.Size())
}

func (f FilesystemStorage) Has(id refs.Blob) (bool, error) {
	_, err := f.Size(id)
	if err != nil {
		if errors.Is(err, replication.ErrBlobNotFound) {
			return false, nil
		}
		return false, errors.Wrap(err, "failed to check blob size")
	}
	return true, nil
}

func (f FilesystemStorage) ensureDirectoriesExist() error {
	if err := f.createStorage(); err != nil {
		return errors.Wrap(err, "failed to create the storage directory")
	}

	if err := f.createTemporary(); err != nil {
		return errors.Wrap(err, "failed to create the temporary directory")
	}

	return nil
}

func (f FilesystemStorage) createStorage() error {
	return os.MkdirAll(f.dirStorage(), onlyForMe)
}

func (f FilesystemStorage) createTemporary() error {
	return os.MkdirAll(f.dirTemporary(), onlyForMe)
}

func (f FilesystemStorage) dirTemporary() string {
	return path.Join(f.path, "tmp")
}

func (f FilesystemStorage) dirStorage() string {
	return path.Join(f.path, "sha256")
}

func (f FilesystemStorage) pathStorage(id refs.Blob) string {
	hexRef := hex.EncodeToString(id.Bytes())
	dirName := hexRef[:charactersInDirName]
	fileName := hexRef[charactersInDirName:]
	return path.Join(f.dirStorage(), dirName, fileName)
}

func (f FilesystemStorage) removeTemporaryFiles() error {
	return filepath.WalkDir(f.dirTemporary(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return errors.Wrap(err, "error removing temporary files")
		}
		if !d.IsDir() {
			if err := os.Remove(path); err != nil {
				return errors.Wrap(err, "could not remove one of the old temporary files")
			}
		}
		return nil
	})
}
