package main

import (
	"io"
	"golang.org/x/net/context"
	"os"
	//"log"
	"runtime"

	"bazil.org/fuse"
)

type BbFsHandle struct {
	os.File
	Node *BbFsNode
}

func NewHandle(node *BbFsNode, f *os.File) *BbFsHandle {
	nh := &BbFsHandle{}
	nh.File = *f
	nh.Node = node
	return nh
}

// ========== methods those implement Handle interfaces ======

func (handle *BbFsHandle) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	logRev(handle.Node.Name, req.String())

	err := handle.Sync()
	if err != nil {
		logRev(err.Error())
		logErr(runtime.Caller(0))
	}
	return err
}

func (handle *BbFsHandle) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	logRev(handle.Node.Name, req.String())

	err := handle.Sync()

	if err != nil {
		logRev(err.Error())
		logErr(runtime.Caller(0))
	}
	return err
}

func (handle *BbFsHandle) ReadDirAll(ctx context.Context) (dirs []fuse.Dirent,
	err error) {
	logRev(handle.Node.Name)
	Log(runtime.Caller(0))

	var names []string
	names, err = handle.Readdirnames(0)

	for _, name := range names {
		attr := &fuse.Attr{}
		npath := handle.Node.newpath(name)
		nfpath := handle.Node.Fs.BbFsFullPath(npath)
		GetfileAttr(nfpath, attr)

		ndir := &fuse.Dirent{}
		ndir.Name = name
		ndir.Inode = attr.Inode
		if ndir.Name == "" {
			ndir.Inode = 2
			ndir.Name = "."
		}
		ndir.Type = fuse.DirentType(attr.Mode >> 12)
		dirs = append(dirs, *ndir)
	}

	return dirs, err
}

func (handle *BbFsHandle) Read(ctx context.Context, req *fuse.ReadRequest,
	resp *fuse.ReadResponse) error {
	logRev(handle.Node.Name, req.String())

	buf := make([]byte, req.Size)

	len, err := handle.ReadAt(buf, req.Offset)

	if err != nil && err != io.EOF {
		logRev(err.Error())
		logErr(runtime.Caller(0))
		return err
	}

	if err == io.EOF && len == 0 {
		logRev(err.Error())
		logErr(runtime.Caller(0))
		return err
	}

	resp.Data = buf[:len]
	return nil
}

func (handle *BbFsHandle) Write(ctx context.Context, req *fuse.WriteRequest,
	resp *fuse.WriteResponse) error {
	logRev(handle.Node.Name, req.String())

	len, err := handle.WriteAt(req.Data, req.Offset)

	if err != nil {
		logRev(err.Error())
		logErr(runtime.Caller(0))
		return err
	}

	resp.Size = len
	return err
}

func (handle *BbFsHandle) Release(ctx context.Context,
	req *fuse.ReleaseRequest) error {
	logRev(handle.Node.Name, req.String())

	err := handle.Close()
	if err != nil {
		logRev(err.Error())
		logErr(runtime.Caller(0))
	}
	return err
}
