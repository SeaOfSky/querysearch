# query search

# how to run
1. 确认 main.go 中的 FilePath 变量为数据文件路径。
2. 在项目根目录运行 `go run main.go`。

# how to visit the server 
example: http://localhost:8199/fuzzysearch?s=woods&f=1&skip=false

- s: 搜索字符
- f: fuzziness
- skip: false是在结果中不过滤google点，true为过滤。
