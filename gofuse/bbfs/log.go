package main

import (
	"log"
	"path"
	"runtime"
)

func logRev(rev ...string) {
	log.Println(rev)
}

func Log(pc uintptr, file string, line int, ok bool) {
	f := runtime.FuncForPC(pc)
	fname := f.Name()
	log.Println("\t", fname)
}

func logErr(pc uintptr, file string, line int, ok bool) {
	f := runtime.FuncForPC(pc)
	fname := f.Name()
	filename := path.Base(file)
	log.Println("\tERR", filename, line, fname, "*********************")
}
