# mindconv

`mindconv` 是一个使用 Go 编写的跨平台命令行工具，用于将 MindManager `.mmap` 思维导图转换为 Markdown 或静态 HTML 文件。

项目不依赖 MindManager，也不需要运行浏览器或后台服务。转换过程全部在本地完成，适合整理会议记录、项目规划、知识结构和其他树状内容。

## 功能

- 读取 ZIP 容器形式的 MindManager `.mmap` 文件
- 解析 `Document.xml` 中的中心主题和子主题层级
- 提取主题备注
- 提取 HTTP、HTTPS、邮件及相对链接
- 输出层级化 Markdown
- 输出带响应式样式的单文件 HTML
- 支持写入指定文件或标准输出
- 默认拒绝覆盖已有文件
- 限制压缩包条目数、XML 大小、深度和节点数
- 过滤 `javascript:` 等不安全链接协议

## 环境要求

- Go 1.22 或更高版本

运行生成的二进制文件不需要安装 Go 或其他运行时。

## 快速开始

进入项目目录：

```bash
cd mindconv
```

不安装二进制文件，直接转换：

```bash
go run ./cmd/mindconv example.mmap
go run ./cmd/mindconv example.mmap -f html
```

`go run` 后的第一个参数必须是 `./cmd/mindconv`，`.mmap` 文件是传递给程序的输入参数，不能直接写成 `go run example.mmap`。

## 构建与安装

### 构建到项目目录

```bash
git clone https://github.com/Weiki886/mindconv.git
cd mindconv
mkdir -p bin
go build -o bin/mindconv ./cmd/mindconv
```

构建完成后，通过相对路径运行：

```bash
./bin/mindconv --help
./bin/mindconv example.mmap -f html
```

仅仅执行 `go build -o bin/mindconv` 不会让系统自动识别 `mindconv` 命令，必须使用 `./bin/mindconv`。

### 安装为全局命令

```bash
go install ./cmd/mindconv
export PATH="$(go env GOPATH)/bin:$PATH"
mindconv --version
```

`go install` 默认将可执行文件安装到 `$(go env GOPATH)/bin`。上面的 `export` 会把该目录临时加入 `PATH`，但只对当前终端窗口有效；关闭窗口后，这项设置就会失效。

在 macOS 默认的 zsh 中，可以按以下步骤永久配置。

1. 确认程序已经安装：

   ```bash
   go install ./cmd/mindconv
   ls "$(go env GOPATH)/bin/mindconv"
   ```

2. 将 Go 的二进制目录写入 zsh 配置文件。以下命令只需执行一次：

   ```bash
   printf '\nexport PATH="$(go env GOPATH)/bin:$PATH"\n' >> ~/.zshrc
   ```

3. 让配置在当前终端立即生效，无需重新打开窗口：

   ```bash
   source ~/.zshrc
   ```

4. 检查终端是否已经能找到 `mindconv`：

   ```bash
   command -v mindconv
   mindconv --version
   ```

`command -v mindconv` 应输出类似 `~/go/bin/mindconv` 的位置，`mindconv --version` 应输出当前版本号。以后新打开的终端会自动读取 `~/.zshrc`，不需要再次执行 `export`。

如果不想修改 shell 配置，也可以始终通过完整安装路径运行：

```bash
"$(go env GOPATH)/bin/mindconv" example.mmap
```

## 卸载

### 删除全局安装的命令

如果通过 `go install ./cmd/mindconv` 安装，先查看终端当前找到的程序位置：

```bash
command -v mindconv
```

Go 会优先把程序安装到 `GOBIN`；没有配置 `GOBIN` 时，则安装到 `$(go env GOPATH)/bin`。可以使用以下命令确定目录并删除 `mindconv`：

```bash
mindconv_bin_dir="$(go env GOBIN)"
if [ -z "$mindconv_bin_dir" ]; then
  mindconv_bin_dir="$(go env GOPATH)/bin"
fi
rm "$mindconv_bin_dir/mindconv"
hash -r
```

删除后进行验证：

```bash
command -v mindconv
```

如果没有任何输出，说明全局命令已经删除。

### 删除项目内构建的程序

如果只执行过 `go build -o bin/mindconv ./cmd/mindconv`，程序仅存在于当前项目的 `bin` 目录：

```bash
rm ./bin/mindconv
```

这种构建方式没有进行全局安装，因此不需要修改 `PATH`。

### 是否需要修改 `~/.zshrc`

通常不需要删除以下配置：

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

其他通过 `go install` 安装的 Go 工具也可能依赖该目录。只有确定以后不再使用任何 Go 全局工具时，才需要从 `~/.zshrc` 删除这一行，然后执行：

```bash
source ~/.zshrc
```

### 删除源代码

卸载命令不会自动删除源码，删除源码也不会自动卸载已经安装的命令。如果不再需要项目代码，可以在完成全局命令卸载后，通过 Finder 将 `mindconv` 文件夹移到废纸篓。

## 使用方法

基本语法：

```text
mindconv [options] <input.mmap>
```

以下示例假设已经通过 `go install` 安装了全局命令。如果使用项目内构建版本，请将 `mindconv` 替换为 `./bin/mindconv`；如果直接运行源码，请替换为 `go run ./cmd/mindconv`。

### 转换为 Markdown

```bash
mindconv project.mmap
```

默认在输入文件旁生成 `project.md`。

### 转换为 HTML

```bash
mindconv --format html project.mmap
```

也可以使用短参数：

```bash
mindconv -f html project.mmap
```

### 指定输出文件

```bash
mindconv project.mmap --output docs/project.md
mindconv project.mmap -f html -o public/project.html
```

### 输出到终端

```bash
mindconv project.mmap --stdout
mindconv project.mmap -f html -o -
```

### 覆盖已有文件

```bash
mindconv project.mmap --force
```

### 查看版本

```bash
mindconv --version
```

## 常见问题

### `zsh: command not found: mindconv`

说明程序还没有安装到 `PATH`。可以使用项目内的二进制：

```bash
./bin/mindconv example.mmap
```

或者按照“安装为全局命令”章节执行 `go install` 并配置 `PATH`。

### `go run example.mmap` 提示 `outside main module`

`go run` 需要接收 Go 程序路径。正确命令是：

```bash
go run ./cmd/mindconv example.mmap
```

### 输出文件已经存在

程序默认保护已有文件。确认需要覆盖时使用：

```bash
mindconv example.mmap --force
```

## 参数

| 参数 | 说明 |
|------|------|
| `-f, --format` | 输出格式，可选 `md`、`markdown`、`html` 或 `htm`；默认 `md` |
| `-o, --output` | 输出路径，使用 `-` 表示标准输出 |
| `--stdout` | 将转换结果输出到终端 |
| `--force` | 覆盖已经存在的输出文件 |
| `--version` | 显示当前版本 |
| `-h, --help` | 显示帮助信息 |

## 转换规则

| MMAP 内容 | Markdown | HTML |
|-----------|----------|------|
| 中心主题 | 一级标题 | 页面标题和根节点 |
| 子主题 | 二至六级标题 | 嵌套主题卡片 |
| 更深层主题 | 缩进列表 | 嵌套主题卡片 |
| 备注 | 标题下方的正文 | 主题备注区域 |
| 超链接 | Markdown 链接 | 自动转义的 HTML 链接 |

## 项目结构

```text
mindconv/
├── cmd/mindconv/       # 命令行入口
├── internal/
│   ├── cli/            # 参数解析、输出和退出码
│   ├── mmap/           # MMAP ZIP/XML 解析
│   ├── model/          # 格式无关的思维导图模型
│   └── render/         # Markdown 与 HTML 渲染
├── go.mod
├── LICENSE
└── README.md
```

## 当前限制

- 目前只支持 ZIP/XML 结构的 `.mmap`，不支持旧版 `.mmp`。
- 暂未还原主题样式、自由布局、边界框和关系线。
- 暂未导出嵌入图片和附件。
- MindManager 不同版本可能存在 XML 结构差异；遇到无法解析的文件时，请提供脱敏后的最小样本和 MindManager 版本信息。
- Markdown 本身无法完整表达思维导图的图形布局，因此输出以内容层级为主。

## 开发与验证

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/mindconv
```

新增格式兼容逻辑时，应在 `internal/mmap/` 增加不包含隐私内容的最小测试样本或合成测试数据。

## 安全说明

`.mmap` 文件可能包含附件或来自不可信来源的数据。`mindconv` 当前不会执行或自动打开附件，并对主要解析资源设置上限。HTML 输出使用 Go 的 `html/template` 自动转义动态内容，但仍建议只在可信环境中打开转换结果。

## 许可证

本项目基于 [MIT License](LICENSE) 发布。
