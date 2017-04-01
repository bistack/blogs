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

type BbFsFile struct {
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

func NewFileObj(fs *BbFs, name string, isDir bool) (nf *BbFsFile) {
	nf = &BbFsFile{}
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

func (file *BbFsFile) newpath(name string) string {
	var npath string
	if file.Name == "" {
		npath = name
	} else {
		npath = file.Name + "/" + name
	}

	return npath
}

// ========== methods those implement fuse interfaces ======

func (file *BbFsFile) Attr(ctx context.Context, attr *fuse.Attr) error {
	logRev(file.Name)
	Log(runtime.Caller(0))

	fpath := file.Fs.BbFsFullPath(file.Name)

	err := GetfileAttr(fpath, attr)

	if file.Name == "" {
		attr.Inode = 1
		attr.Mode = os.ModeDir | 0755
	}

	if file.IsDir {
		attr.Mode |= os.ModeDir
	}

	return err
}

/*
func (file *BbFsFile) Getattr(ctx context.Context, req *fuse.GetattrRequest,
	resp *fuse.GetattrResponse) error {
	logRev(file.Name, req.String())

	fpath := file.Fs.BbFsFullPath(file.Name)

	err := GetfileAttr(fpath, &resp.Attr)
	if err != nil {
		resp.Attr.Inode = file.Inode
		logErr(runtime.Caller(0))
	}

	return err
}
*/

func (file *BbFsFile) Setattr(ctx context.Context, req *fuse.SetattrRequest,
	resp *fuse.SetattrResponse) error {
	logRev(file.Name, req.String())
	Log(runtime.Caller(0))

	fpath := file.Fs.BbFsFullPath(file.Name)

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

	if err == nil && req.Valid.Handle() && (req.Handle > 0) {
		logErr(runtime.Caller(0))
	}

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

func (file *BbFsFile) Symlink(ctx context.Context,
	req *fuse.SymlinkRequest) (fs.Node, error) {
	logRev(file.Name, req.String())
	Log(runtime.Caller(0))

	fname := file.newpath(req.NewName)
	flink := file.Fs.BbFsFullPath(fname)

	fpath := file.Fs.BbFsFullPath(req.Target)
	err := os.Symlink(fpath, flink)

	var nlink *BbFsFile = nil
	if err == nil {
		nlink = NewFileObj(file.Fs, fname, file.IsDir)
	} else {
		logErr(runtime.Caller(0))
	}
	return nlink, err
}

func (file *BbFsFile) Readlink(ctx context.Context,
	req *fuse.ReadlinkRequest) (string, error) {
	logRev(file.Name, req.String())
	Log(runtime.Caller(0))

	fpath := file.Fs.BbFsFullPath(file.Name)
	tgt, err := os.Readlink(fpath)

	if err != nil {
		logErr(runtime.Caller(0))
	}
	return tgt, err
}

func (file *BbFsFile) Link(ctx context.Context,
	req *fuse.LinkRequest, old fs.Node) (fs.Node, error) {
	logRev(file.Name, req.String())
	Log(runtime.Caller(0))

	if !file.IsDir {
		return nil, fuse.ENOTSUP
	}

	fpath := file.Fs.BbFsFullPath(file.Name)
	flink := file.Fs.BbFsFullPath(req.NewName)

	err := os.Link(fpath, flink)

	var link *BbFsFile = nil
	if err == nil {
		link = NewFileObj(file.Fs, req.NewName, file.IsDir)
	} else {
		logErr(runtime.Caller(0))
	}
	return link, err
}

func (file *BbFsFile) Remove(ctx context.Context, req *fuse.RemoveRequest) (err error) {
	logRev(file.Name, req.String())
	Log(runtime.Caller(0))

	if !file.IsDir {
		return fuse.ENOSYS
	}

	fpath := file.Fs.BbFsFullPath(req.Name)

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

func (file *BbFsFile) Access(ctx context.Context, req *fuse.AccessRequest) error {
	//logRev(file.Name, req.String())
	return nil
}

func (file *BbFsFile) Lookup(ctx context.Context, name string) (fs.Node, error) {
	logRev(file.Name, name)
	Log(runtime.Caller(0))

	if !file.IsDir {
		logErr(runtime.Caller(0))
		return nil, fuse.ENOENT
	}

	fpath := file.Fs.BbFsFullPath(file.Name)
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

	var ffile *BbFsFile = nil
	if found {
		npath := file.newpath(name)

		fpath := file.Fs.BbFsFullPath(npath)

		var fi os.FileInfo
		fi, err = os.Lstat(fpath)

		if err != nil {
			logErr(runtime.Caller(0))
		}

		ffile = NewFileObj(file.Fs, npath, fi.IsDir())
	} else {
		err = fuse.ENOENT
	}

	return ffile, err
}

/*
func (file *BbFsFile) Lookup(ctx context.Context, req *fuse.LookupRequest,
	resp *fuse.LookupResponse) (fs.Node, error) {
	logRev(file.Name, req.String())
	Log(runtime.Caller(0))
	return nil, fuse.ENOTSUP
}
*/
func (file *BbFsFile) Mkdir(ctx context.Context,
	req *fuse.MkdirRequest) (fs.Node, error) {
	logRev(file.Name, req.String())

	if !file.IsDir {
		logErr(runtime.Caller(0))
		return nil, fuse.ENOSYS
	}

	npath := file.newpath(req.Name)

	fpath := file.Fs.BbFsFullPath(npath)

	err := os.MkdirAll(fpath, req.Mode)

	var ndir *BbFsFile = nil
	if err == nil {
		ndir = NewFileObj(file.Fs, npath, true)
	} else {
		logErr(runtime.Caller(0))
	}
	return ndir, err
}

func (file *BbFsFile) Open(ctx context.Context, req *fuse.OpenRequest,
	resp *fuse.OpenResponse) (fs.Handle, error) {
	logRev(file.Name, req.String())

	// TODO
	fpath := file.Fs.BbFsFullPath(file.Name)
	f, err := os.OpenFile(fpath, int(req.Flags), 644)

	if err != nil {
		logErr(runtime.Caller(0))
		return nil, err
	}

	resp.Handle = fuse.HandleID(f.Fd())
	resp.Flags = fuse.OpenResponseFlags(req.Flags)

	return file, err
}

func (file *BbFsFile) Create(ctx context.Context, req *fuse.CreateRequest,
	resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	logRev(file.Name, req.String())

	npath := file.newpath(req.Name)
	fpath := file.Fs.BbFsFullPath(npath)

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

	nfile := NewFileObj(file.Fs, npath, isDir)

	return nfile, nfile, err
}

func (file *BbFsFile) Forget() {
	logRev(file.Name)
	Log(runtime.Caller(0))
	return
}

func (file *BbFsFile) Rename(ctx context.Context, req *fuse.RenameRequest,
	newDir fs.Node) error {
	logRev(file.Name, req.String())

	fpath := file.Fs.BbFsFullPath(req.OldName)
	npath := file.Fs.BbFsFullPath(req.NewName)

	err := os.Rename(fpath, npath)

	if err != nil {
		logErr(runtime.Caller(0))
	}
	return err
}

func (file *BbFsFile) Mknod(ctx context.Context,
	req *fuse.MknodRequest) (fs.Node, error) {
	logRev(file.Name, req.String())

	fpath := file.Fs.BbFsFullPath(req.Name)
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

	var nfile *BbFsFile = nil

	if err == nil {
		nfile = NewFileObj(file.Fs, req.Name, false)
	} else {
		logErr(runtime.Caller(0))
	}
	return nfile, err
}

func (file *BbFsFile) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	logRev(file.Name, req.String())
	fd := uintptr(req.Handle)
	f := os.NewFile(fd, file.Name)
	err := f.Sync()
	if err != nil {
		logErr(runtime.Caller(0))
	}
	return nil
}

func (file *BbFsFile) Getxattr(ctx context.Context, req *fuse.GetxattrRequest,
	resp *fuse.GetxattrResponse) error {
	logRev(file.Name, req.String())
	return nil
}

func (file *BbFsFile) Listxattr(ctx context.Context, req *fuse.ListxattrRequest,
	resp *fuse.ListxattrResponse) error {
	logRev(file.Name, req.String())
	return nil
}

func (file *BbFsFile) Setxattr(ctx context.Context,
	req *fuse.SetxattrRequest) error {
	logRev(file.Name, req.String())
	return nil
}

func (file *BbFsFile) Removexattr(ctx context.Context,
	req *fuse.RemovexattrRequest) error {
	logRev(file.Name, req.String())
	return nil
}

func (file *BbFsFile) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	logRev(file.Name, req.String())
	fd := uintptr(req.Handle)
	f := os.NewFile(fd, file.Name)
	err := f.Sync()

	if err != nil {
		logRev(err.Error())
		logErr(runtime.Caller(0))
	}
	return err
}
/*
func (file *BbFsFile) ReadAll(ctx context.Context) ([]byte, error) {
	logRev(file.Name)
	Log(runtime.Caller(0))

	fpath := file.Fs.BbFsFullPath(file.Name)

	fi, err := os.Stat(fpath)

	len := 0
	max_len := 512*1024
	if fi.Size() > max_len {
		len = max_len
	} else {
		len = fi.Size()
	}
	buf := make([]byte, len)
	var f *os.File
	f, err = os.Open(fpath)
	_, err = f.Read(buf)
	err = f.Close()

	return buf, err
}
*/
func (file *BbFsFile) ReadDirAll(ctx context.Context) (dirs []fuse.Dirent,
	err error) {
	logRev(file.Name)
	Log(runtime.Caller(0))

	if !file.IsDir {
		logErr(runtime.Caller(0))
		return nil, fuse.ENOENT
	}

	fpath := file.Fs.BbFsFullPath(file.Name)
	f, err := os.Open(fpath)

	if err != nil {
		logErr(runtime.Caller(0))
		return nil, err
	}

	var names []string
	names, err = f.Readdirnames(0)
	err = f.Close()

	for _, name := range names {
		attr := &fuse.Attr{}
		npath := file.newpath(name)
		nfpath := file.Fs.BbFsFullPath(npath)
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

func (file *BbFsFile) Read(ctx context.Context, req *fuse.ReadRequest,
	resp *fuse.ReadResponse) error {
	logRev(file.Name, req.String())

	fd := uintptr(req.Handle)
	f := os.NewFile(fd, file.Name)

	var len int
	buf := make([]byte, req.Size)

/*
	_, err := f.Seek(req.Offset, 0)
	if err != nil {
		logRev(err.Error())
		logErr(runtime.Caller(0))
		return err
	}
	
	len, err := f.Read(buf)
*/
	len, err := f.ReadAt(buf, req.Offset)

	if err != nil {
		logRev(err.Error())
		logErr(runtime.Caller(0))
		return err
	}

	resp.Data = buf[:len]
	return err
}

func (file *BbFsFile) Write(ctx context.Context, req *fuse.WriteRequest,
	resp *fuse.WriteResponse) error {
	logRev(file.Name, req.String())

	fd := uintptr(req.Handle)
	f := os.NewFile(fd, file.Name)
/*
	_, err := f.Seek(req.Offset, 0)
	if err != nil {
		logRev(err.Error())
		logErr(runtime.Caller(0))
		return err
	}

	len, err := f.Write(req.Data)
*/
	len, err := f.WriteAt(req.Data, req.Offset)

	if err != nil {
		logRev(err.Error())
		logErr(runtime.Caller(0))
		return err
	}

	resp.Size = len
	return err
}

func (file *BbFsFile) Release(ctx context.Context,
	req *fuse.ReleaseRequest) error {
	logRev(file.Name, req.String())

	fd := uintptr(req.Handle)
	f := os.NewFile(fd, file.Name)
	err := f.Close()
	if err != nil {
		logRev(err.Error())
		logErr(runtime.Caller(0))
	}
	return err
}
