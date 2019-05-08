# giant-parrot
Explore the citation graph of an article published on Pubmed.

Giant parrot in the sense of "What is standing on the shoulder of giants".

## Build

```{bash}
# In src/
go generate
go build -ldflags "-H windowsgui" -o giant-parrot.exe
```