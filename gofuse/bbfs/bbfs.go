package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Useage:\n\t%s ROOT_DIR MNT_DIR\n", os.Args[0])

	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		return
	}

	var rootDir string
	var mntDir string
	rootDir = flag.Arg(0)
	mntDir = flag.Arg(1)

	// get a mount connection
	mntConn, mnterr := fuse.Mount(mntDir,
		fuse.FSName("bbfs"),
		fuse.Subtype("gobbFs"),
		fuse.LocalVolume(),
		fuse.VolumeName("bbFs"),
	)

	defer func() {
		if mnterr == nil && mntConn != nil {
			mntConn.Close()
		}
	}()

	if mnterr != nil {
		logErr(runtime.Caller(0))
		return
	}

	// wait mount
	<-mntConn.Ready

	var err error = nil
	err = mntConn.MountError

	if err != nil {
		logErr(runtime.Caller(0))
		return
	}
	
	bbFs := NewBbFs(rootDir)

	Log(runtime.Caller(0))
	// run a go-fuse server
	go func() {
		err = fs.Serve(mntConn, bbFs)
		if err != nil {
			logErr(runtime.Caller(0))
			bbFs.Destroy()
			return
		}
	}()

	// wait umount
	<-bbFs.Umnt
}
