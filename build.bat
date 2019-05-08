@echo off
cd src\
go generate
go build -ldflags "-H windowsgui" -o giant-parrot.exe
cd ..