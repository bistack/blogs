package main

import (
	"golang.org/x/net/context"
	"path/filepath"
	"runtime"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
)

type BbFs struct {
	root string
	Umnt chan bool
}

func NewBbFs(rootDir string) *BbFs {
	fs := &BbFs{}
	fs.Umnt = make(chan bool)
	var err error = nil
	fs.root, err = filepath.Abs(rootDir)
	if err != nil {
		logErr(runtime.Caller(0))
	}

	return fs
}

func (fs *BbFs) BbFsFullPath(path string) string {
	if path == "" {
		return fs.root
	}
	fpath := fs.root + "/" + path
	return fpath
}

// =========== methods those implement fuse interfaces ===

func (fs *BbFs) Root() (fs.Node, error) {
	logRev("FS")
	Log(runtime.Caller(0))
	root := NewFileObj(fs, "", true)

	return root, nil
}

func respStatfs(resp *fuse.StatfsResponse, statfs *syscall.Statfs_t) {
	resp.Blocks = statfs.Blocks
	resp.Bfree = statfs.Bfree
	resp.Bavail = statfs.Bavail
	resp.Files = statfs.Files
	resp.Ffree = statfs.Ffree
	resp.Bsize = uint32(statfs.Bsize)
	resp.Namelen = 512
	resp.Frsize = 0
}

func (fs *BbFs) Statfs(ctx context.Context, req *fuse.StatfsRequest,
	resp *fuse.StatfsResponse) error {
	logRev("FS", req.String())

	statfs := &syscall.Statfs_t{}

	err := syscall.Statfs(fs.root, statfs)

	if err == nil {
		respStatfs(resp, statfs)
	} else {
		logErr(runtime.Caller(0))
	}

	return err
}

func (fs *BbFs) Destroy() {
	logRev("FS")
	Log(runtime.Caller(0))

	fs.Umnt <- true
	return
}

func (fs *BbFs) GenerateInode(parentInode uint64, name string) uint64 {
	logRev("FS")
	Log(runtime.Caller(0))

	var i uint64 = 0

	return func() uint64 {
		i += 1
		return i
	}()
}
