package hotline

import (
	"github.com/stretchr/testify/mock"
	"io/fs"
	"os"
)

type FileStore interface {
	Mkdir(name string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
	Open(name string) (*os.File, error)
	Symlink(oldname, newname string) error
	Remove(name string) error
	Create(name string) (*os.File, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	// TODO: implement these
	// Rename(oldpath string, newpath string) error
	// RemoveAll(path string) error
}

type OSFileStore struct{}

func (fs *OSFileStore) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (fs *OSFileStore) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (fs *OSFileStore) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (fs *OSFileStore) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func (fs *OSFileStore) Remove(name string) error {
	return os.Remove(name)
}

func (fs *OSFileStore) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (fs *OSFileStore) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

type MockFileStore struct {
	mock.Mock
}

func (mfs *MockFileStore) Mkdir(name string, perm os.FileMode) error {
	args := mfs.Called(name, perm)
	return args.Error(0)
}

func (mfs *MockFileStore) Stat(name string) (os.FileInfo, error) {
	args := mfs.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)

	}
	return args.Get(0).(os.FileInfo), args.Error(1)
}

func (mfs *MockFileStore) Open(name string) (*os.File, error) {
	args := mfs.Called(name)
	return args.Get(0).(*os.File), args.Error(1)
}

func (mfs *MockFileStore) Symlink(oldname, newname string) error {
	args := mfs.Called(oldname, newname)
	return args.Error(0)
}

func (mfs *MockFileStore) Remove(name string) error {
	args := mfs.Called(name)
	return args.Error(0)
}

func (mfs *MockFileStore) Create(name string) (*os.File, error) {
	args := mfs.Called(name)
	return args.Get(0).(*os.File), args.Error(1)
}

func (mfs *MockFileStore) WriteFile(name string, data []byte, perm fs.FileMode) error {
	args := mfs.Called(name, data, perm)
	return args.Error(0)
}
