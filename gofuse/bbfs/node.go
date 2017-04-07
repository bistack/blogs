package main

import (
	"golang.org/x/net/context"
	"os"
	"runtime"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type BbFsNode struct {
	Fs    *BbFs
	Name  string
	IsDir bool
}

func setTvTime(tv *syscall.Timeval, time time.Time) {
	tv.Sec = time.Unix()
	//tv.Usec = int64(time.Nanosecond() / 1000)
	tv.Usec = int32(time.Nanosecond() / 1000)
}

func timeTs(ts syscall.Timespec) time.Time {
	return time.Unix(ts.Sec, ts.Nsec)
}

func fileMode(unixMode uint16) os.FileMode {
	mode := os.FileMode(unixMode & 0777)
	switch unixMode & syscall.S_IFMT {
	case syscall.S_IFREG:
		// do nothing
	case syscall.S_IFDIR:
		mode |= os.ModeDir
	case syscall.S_IFCHR:
		mode |= os.ModeCharDevice | os.ModeDevice
	case syscall.S_IFBLK:
		mode |= os.ModeDevice
	case syscall.S_IFIFO:
		mode |= os.ModeNamedPipe
	case syscall.S_IFLNK:
		mode |= os.ModeSymlink
	case syscall.S_IFSOCK:
		mode |= os.ModeSocket
	default:
		mode |= os.ModeDevice
	}
	if unixMode&syscall.S_ISUID != 0 {
		mode |= os.ModeSetuid
	}
	if unixMode&syscall.S_ISGID != 0 {
		mode |= os.ModeSetgid
	}
	return mode
}

func attrStat(attr *fuse.Attr, stat *syscall.Stat_t) {
	attr.Valid = 10
	attr.Inode = stat.Ino
	attr.Size = uint64(stat.Size)
	attr.Blocks = uint64(stat.Blocks)
//	attr.Atime = timeTs(stat.Atimespec)
//	attr.Ctime = timeTs(stat.Ctimespec)
//	attr.Crtime = timeTs(stat.Birthtimespec)
	attr.Mode = fileMode(uint16(stat.Mode))
	attr.Nlink = uint32(stat.Nlink)
	attr.Uid = stat.Uid
	attr.Gid = stat.Gid
	attr.Rdev = uint32(stat.Rdev)
//	attr.Flags = stat.Flags
	attr.BlockSize = uint32(stat.Blksize)
}

func NewNode(fs *BbFs, name string, isDir bool) (nf *BbFsNode) {
	nf = &BbFsNode{}
	nf.Name = name
	nf.Fs = fs
	nf.IsDir = isDir
	return nf
}

func GetfileAttr(fpath string, attr *fuse.Attr) error {
	stat := &syscall.Stat_t{}

	err := syscall.Lstat(fpath, stat)

	if err == nil {
		attrStat(attr, stat)
	} else {
		logErr(runtime.Caller(0))
	}

	return err
}

func (node *BbFsNode) newpath(name string) string {
	var npath string
	if node.Name == "" {
		npath = name
	} else {
		npath = node.Name + "/" + name
	}

	return npath
}

// ========== methods those implement Node interfaces ======

func (node *BbFsNode) Attr(ctx context.Context, attr *fuse.Attr) error {
	logRev(node.Name)
	Log(runtime.Caller(0))

	fpath := node.Fs.BbFsFullPath(node.Name)

	err := GetfileAttr(fpath, attr)

	if err != nil {
		logErr(runtime.Caller(0))
		return err
	}

	if node.Name == "" {
		attr.Inode = 1
		attr.Mode = os.ModeDir | 0755
	}

	if node.IsDir {
		attr.Mode |= os.ModeDir
	}

	return nil
}

func (node *BbFsNode) Setattr(ctx context.Context, req *fuse.SetattrRequest,
	resp *fuse.SetattrResponse) error {
	logRev(node.Name, req.String())
	Log(runtime.Caller(0))

	fpath := node.Fs.BbFsFullPath(node.Name)

	var err error
	err = nil
	if err == nil && req.Valid.Mode() {
		err = os.Chmod(fpath, req.Mode)
		if err == nil {
			resp.Attr.Mode = req.Mode
		} else {
			logErr(runtime.Caller(0))
		}
	}

	if err == nil && (req.Valid.Gid() && req.Valid.Uid()) {
		err = os.Chown(fpath, int(req.Uid), int(req.Gid))
		if err == nil {
			resp.Attr.Gid = req.Gid
			resp.Attr.Uid = req.Uid
		} else {
			logErr(runtime.Caller(0))
		}
	}

	if err == nil && req.Valid.Size() {
		err = os.Truncate(fpath, int64(req.Size))
		if err == nil {
			resp.Attr.Size = req.Size
		} else {
			logErr(runtime.Caller(0))
		}
	}

	var amtv [2]syscall.Timeval
	if err == nil && (req.Valid.Atime() || req.Valid.AtimeNow()) {
		setTvTime(&amtv[0], req.Atime)
		err = syscall.Utimes(fpath, amtv[0:2])
		if err == nil {
			resp.Attr.Atime = req.Atime
		} else {
			logErr(runtime.Caller(0))
		}
	}

	if err == nil && (req.Valid.Mtime() || req.Valid.MtimeNow()) {
		setTvTime(&amtv[1], req.Mtime)
		err = syscall.Utimes(fpath, amtv[0:2])
		if err == nil {
			resp.Attr.Mtime = req.Mtime
		} else {
			logErr(runtime.Caller(0))
		}
	}

/*
	if err == nil && req.Valid.Handle() {
		logErr(runtime.Caller(0))
	}
*/
	// OS X only
/*
	if err == nil && req.Valid.Flags() {
		err = syscall.Chflags(fpath, int(req.Flags))
		if err == nil {
			resp.Attr.Flags = req.Flags
		} else {
			logErr(runtime.Caller(0))
		}
	}
*/
	if err != nil {
		logErr(runtime.Caller(0))
	}
	return err
}

func (node *BbFsNode) Symlink(ctx context.Context,
	req *fuse.SymlinkRequest) (fs.Node, error) {
	logRev(node.Name, req.String())
	Log(runtime.Caller(0))

	fname := node.newpath(req.NewName)
	flink := node.Fs.BbFsFullPath(fname)

	fpath := node.Fs.BbFsFullPath(req.Target)
	err := os.Symlink(fpath, flink)

	if err != nil {
		logErr(runtime.Caller(0))
		return nil, err
	}

	nlink := NewNode(node.Fs, fname, node.IsDir)
	return nlink, nil
}

func (node *BbFsNode) Readlink(ctx context.Context,
	req *fuse.ReadlinkRequest) (string, error) {
	logRev(node.Name, req.String())
	Log(runtime.Caller(0))

	fpath := node.Fs.BbFsFullPath(node.Name)
	tgt, err := os.Readlink(fpath)

	if err != nil {
		logErr(runtime.Caller(0))
	}
	return tgt, err
}

func (node *BbFsNode) Link(ctx context.Context,
	req *fuse.LinkRequest, old fs.Node) (fs.Node, error) {
	logRev(node.Name, req.String())
	Log(runtime.Caller(0))

	if !node.IsDir {
		return nil, fuse.ENOTSUP
	}

	fpath := node.Fs.BbFsFullPath(node.Name)
	flink := node.Fs.BbFsFullPath(req.NewName)

	err := os.Link(fpath, flink)

	if err != nil {
		logErr(runtime.Caller(0))
		return nil, err
	}

	link := NewNode(node.Fs, req.NewName, node.IsDir)
	return link, nil
}

func (node *BbFsNode) Remove(ctx context.Context, req *fuse.RemoveRequest) (err error) {
	logRev(node.Name, req.String())
	Log(runtime.Caller(0))

	if !node.IsDir {
		return fuse.ENOSYS
	}

	fpath := node.Fs.BbFsFullPath(req.Name)

	if req.Dir {
		err = os.RemoveAll(fpath)
	} else {
		err = os.Remove(fpath)
	}

	if err != nil {
		logErr(runtime.Caller(0))
	}
	return err
}

func (node *BbFsNode) Access(ctx context.Context, req *fuse.AccessRequest) error {
	//logRev(node.Name, req.String())
	return nil
}

func (node *BbFsNode) Lookup(ctx context.Context, name string) (fs.Node, error) {
	logRev(node.Name, name)
	Log(runtime.Caller(0))

	if !node.IsDir {
		logErr(runtime.Caller(0))
		return nil, fuse.ENOENT
	}

	fpath := node.Fs.BbFsFullPath(node.Name)
	f, err := os.Open(fpath)

	if err != nil {
		logErr(runtime.Caller(0))
		return nil, err
	}

	var names []string
	var found bool = false
	names, err = f.Readdirnames(128)
	for err == nil {
		for _, item := range names {
			if item == name {
				found = true
				break
			}
		}

		if found {
			break
		}

		names, err = f.Readdirnames(128)
	}
	err = f.Close()

	var fnode *BbFsNode = nil
	if found {
		npath := node.newpath(name)

		fpath := node.Fs.BbFsFullPath(npath)

		var fi os.FileInfo
		fi, err = os.Lstat(fpath)

		if err != nil {
			logErr(runtime.Caller(0))
		}

		fnode = NewNode(node.Fs, npath, fi.IsDir())
	} else {
		err = fuse.ENOENT
	}

	return fnode, err
}

func (node *BbFsNode) Mkdir(ctx context.Context,
	req *fuse.MkdirRequest) (fs.Node, error) {
	logRev(node.Name, req.String())

	if !node.IsDir {
		logErr(runtime.Caller(0))
		return nil, fuse.ENOSYS
	}

	npath := node.newpath(req.Name)

	fpath := node.Fs.BbFsFullPath(npath)

	err := os.MkdirAll(fpath, req.Mode)

	if err != nil {
		logErr(runtime.Caller(0))
		return nil, err
	}

	ndir := NewNode(node.Fs, npath, true)
	return ndir, nil
}

func (node *BbFsNode) Open(ctx context.Context, req *fuse.OpenRequest,
	resp *fuse.OpenResponse) (fs.Handle, error) {
	logRev(node.Name, req.String())

	// TODO
	fpath := node.Fs.BbFsFullPath(node.Name)
	f, err := os.OpenFile(fpath, int(req.Flags), 644)

	if err != nil {
		logErr(runtime.Caller(0))
		return nil, err
	}

	resp.Handle = fuse.HandleID(f.Fd())
	resp.Flags = fuse.OpenResponseFlags(req.Flags)

	nhandle := NewHandle(node, f)

	return nhandle, nil
}

func (node *BbFsNode) Create(ctx context.Context, req *fuse.CreateRequest,
	resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	logRev(node.Name, req.String())

	npath := node.newpath(req.Name)
	fpath := node.Fs.BbFsFullPath(npath)

	f, err := os.OpenFile(fpath, int(req.Flags), 0644)

	if err != nil {
		logErr(runtime.Caller(0))
		return nil, nil, err
	}

	err = GetfileAttr(fpath, &resp.Attr)

	if err != nil {
		logErr(runtime.Caller(0))
		return nil, nil, err
	}

	resp.Node = fuse.NodeID(resp.Attr.Inode)
	resp.Generation = 0
	resp.EntryValid = 10

	resp.Handle = fuse.HandleID(uint64(f.Fd()))
	resp.Flags = fuse.OpenResponseFlags(req.Flags)

	isDir := false
	if req.Mode&os.ModeDir != 0 {
		isDir = true
	}

	nfile := NewNode(node.Fs, npath, isDir)
	nhandle := NewHandle(nfile, f)
	return nfile, nhandle, err
}

func (node *BbFsNode) Forget() {
	logRev(node.Name)
	Log(runtime.Caller(0))
	return
}

func (node *BbFsNode) Rename(ctx context.Context, req *fuse.RenameRequest,
	newDir fs.Node) error {
	logRev(node.Name, req.String())

	fpath := node.Fs.BbFsFullPath(req.OldName)
	npath := node.Fs.BbFsFullPath(req.NewName)

	err := os.Rename(fpath, npath)

	if err != nil {
		logErr(runtime.Caller(0))
	}
	return err
}

func (node *BbFsNode) Mknod(ctx context.Context,
	req *fuse.MknodRequest) (fs.Node, error) {
	logRev(node.Name, req.String())

	fpath := node.Fs.BbFsFullPath(req.Name)
	var err error
	switch req.Mode {
	case syscall.S_IFREG:
		var f *os.File
		f, err = os.OpenFile(fpath, int(req.Mode),
			syscall.O_CREAT|syscall.O_EXCL|syscall.O_WRONLY)
		if f.Fd() > 0 {
			err = f.Close()
		}
	case syscall.S_IFIFO:
		err = syscall.Mkfifo(fpath, uint32(req.Mode))
	case syscall.S_IFCHR:
		fallthrough
	case syscall.S_IFBLK:
		err = syscall.Mknod(fpath, uint32(req.Mode), int(req.Rdev))
	default:
		err = error(syscall.EINVAL)
	}

	if err != nil {
		logErr(runtime.Caller(0))
		return nil, err
	}

	nnode := NewNode(node.Fs, req.Name, false)
	return nnode, nil
}

func (node *BbFsNode) Getxattr(ctx context.Context, req *fuse.GetxattrRequest,
	resp *fuse.GetxattrResponse) error {
	logRev(node.Name, req.String())
	return nil
}

func (node *BbFsNode) Listxattr(ctx context.Context, req *fuse.ListxattrRequest,
	resp *fuse.ListxattrResponse) error {
	logRev(node.Name, req.String())
	return nil
}

func (node *BbFsNode) Setxattr(ctx context.Context,
	req *fuse.SetxattrRequest) error {
	logRev(node.Name, req.String())
	return nil
}

func (node *BbFsNode) Removexattr(ctx context.Context,
	req *fuse.RemovexattrRequest) error {
	logRev(node.Name, req.String())
	return nil
}
