package webui
import("embed";"io/fs")
//go:embed dist/* dist/assets/*
var files embed.FS
func Files() fs.FS { f,e:=fs.Sub(files,"dist"); if e!=nil {panic(e)};return f }
